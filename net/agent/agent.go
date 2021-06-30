package cherryAgent

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"net"
	"sync/atomic"
	"time"
)

const (
	agentWriteBacklog = 64
)

const (
	Init = iota
	WaitAck
	Working
	Closed
)

type (
	RPCHandler func(session *cherrySession.Session, msg *cherryMessage.Message)

	// process packet listener function
	PacketListener func(agent *Agent, packet cherryFacade.IPacket)

	Options struct {
		Heartbeat       time.Duration                        // heartbeat(sec)
		DataCompression bool                                 // data compression
		PacketListener  map[cherryPacket.Type]PacketListener // process packet listener function
		RPCHandler      RPCHandler                           // rpc handler
	}

	Agent struct {
		Options
		app     cherryFacade.IApplication
		Session *cherrySession.Session // session
		conn    cherryFacade.INetConn  // low-level conn fd
		state   int32                  // current session state
		chDie   chan bool              // wait for close
		chSend  chan pendingMessage    // push message queue
		chWrite chan []byte            // push bytes queue
		lastAt  int64                  // last heartbeat unix time stamp
	}

	pendingMessage struct {
		typ     cherryMessage.Type // message type
		route   string             // message route(push)
		mid     uint64             // response message id(response)
		payload interface{}        // payload
	}
)

func (p *pendingMessage) String() string {
	return fmt.Sprintf("typ = %d, route = %s, mid =%d, payload=%v", p.typ, p.route, p.mid, p.payload)
}

func NewAgent(app cherryFacade.IApplication, opt Options, conn cherryFacade.INetConn) *Agent {
	agent := &Agent{
		app:     app,
		Options: opt,
		conn:    conn,
		state:   Init,
		chDie:   make(chan bool),
		chSend:  make(chan pendingMessage, agentWriteBacklog),
		chWrite: make(chan []byte, agentWriteBacklog),
		lastAt:  time.Now().Unix(),
	}

	return agent
}

func (a *Agent) Status() int32 {
	return atomic.LoadInt32(&a.state)
}

func (a *Agent) SetStatus(state int32) {
	atomic.StoreInt32(&a.state, state)
}

func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, time.Now().Unix())
}

func (a *Agent) SendRaw(bytes []byte) error {
	a.chWrite <- bytes
	return nil
}

func (a *Agent) Send(typ cherryMessage.Type, route string, mid uint64, v interface{}) (err error) {
	if a.Status() == Closed {
		return cherryError.ClusterBrokenPipe
	}

	if len(a.chSend) >= agentWriteBacklog {
		return cherryError.ClusterBufferExceed
	}

	p := pendingMessage{typ: typ, mid: mid, route: route, payload: v}

	a.chSend <- p

	return nil
}

// Push, implementation for session.NetworkEntity interface
func (a *Agent) Push(route string, val interface{}) error {
	return a.Send(cherryMessage.TypePush, route, 0, val)
}

// RPC, implementation for session.NetworkEntity interface
func (a *Agent) RPC(route string, val interface{}) error {
	if a.Status() == Closed {
		return cherryError.ClusterBrokenPipe
	}

	if a.RPCHandler == nil {
		return cherryError.ClusterRPCHandleNotFound
	}

	data, err := a.app.Marshal(val)
	if err != nil {
		return err
	}

	msg := &cherryMessage.Message{
		Type:  cherryMessage.TypeNotify,
		Route: route,
		Data:  data,
	}

	a.RPCHandler(a.Session, msg)

	return nil
}

// Response, implementation for session.NetworkEntity interface
// Response message to session
func (a *Agent) Response(mid uint64, v interface{}) error {
	return a.Send(cherryMessage.TypeResponse, "", mid, v)
}

// Kick
func (a *Agent) Kick(reason string) error {
	pkg, err := a.app.PacketEncode(byte(cherryPacket.Kick), nil)
	if err != nil {
		return err
	}

	_, err = a.conn.Write(pkg)
	if err != nil {
		cherryLogger.Warn(err)
	}

	a.Session.Debugf("kick session, reason[%s]", reason)

	return nil
}

// Close closes the Agent, clean inner state and close low-level connection.
func (a *Agent) Close() {
	if a.Status() == Closed {
		return
	}

	a.SetStatus(Closed)

	if err := a.conn.Close(); err != nil {
		a.Session.Debugf("session close error. [%s]", err)
	}

	a.chDie <- true
}

// RemoteAddr, implementation for session.NetworkEntity interface
// returns the remote network address.
func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

// String, implementation for Stringer interface
func (a *Agent) String() string {
	return fmt.Sprintf("addr=%s, lastTime=%d",
		a.conn.RemoteAddr().String(),
		atomic.LoadInt64(&a.lastAt),
	)
}

func (a *Agent) Run() {
	go a.read()
	go a.write()
}

func (a *Agent) read() {
	defer func() {
		a.Session.Closed()
	}()

	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			a.Session.Debugf("will be closed immediately. error[%s]", err.Error())
			return
		}

		packets, err := a.app.PacketDecode(msg)
		if err != nil {
			cherryLogger.Warnf("packet decoder error. error[%s], msg[%s]", err, msg)
			continue
		}

		if len(packets) < 1 {
			continue
		}

		for _, packet := range packets {
			a.processPacket(packet)
		}
	}
}

func (a *Agent) processPacket(packet cherryFacade.IPacket) {

	listener, found := a.PacketListener[cherryPacket.Type(packet.Type())]
	if found == false {
		a.Session.Debugf("packet[%s] not found.", packet)
		return
	}

	listener(a, packet)
	// update last time
	a.SetLastAt()
}

func (a *Agent) write() {
	ticker := time.NewTicker(a.Heartbeat)
	defer func() {
		ticker.Stop()

		close(a.chSend)
		close(a.chWrite)
		close(a.chDie)
	}()

	for {
		select {
		case <-a.chDie:
			return

		case <-ticker.C:
			deadline := time.Now().Add(-a.Heartbeat).Unix()
			if a.lastAt < deadline {
				a.Session.Debug("connect heartbeat timeout.")
				return
			}

		case bytes := <-a.chWrite:
			_, err := a.conn.Write(bytes)
			if err != nil {
				cherryLogger.Warn(err)
				return
			}

		case data := <-a.chSend:
			payload, err := a.app.Marshal(data.payload)
			if err != nil {
				a.Session.Debug("message serializer error. data[%s]", data.String())
				return
			}

			// construct message and encode
			m := &cherryMessage.Message{
				Type:  data.typ,
				ID:    data.mid,
				Route: data.route,
				Data:  payload,
			}

			// encode message
			em, err := cherryMessage.Encode(m)
			if err != nil {
				cherryLogger.Warn(err)
				break
			}

			// encode packet
			p, err := a.app.PacketEncode(byte(cherryPacket.Data), em)
			if err != nil {
				cherryLogger.Warn(err)
				break
			}
			a.chWrite <- p
		}
	}
}
