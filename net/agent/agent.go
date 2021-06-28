package cherryAgent

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/packet"
	"github.com/cherry-game/cherry/net/session"
	"github.com/cherry-game/cherry/profile"
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
	SessionListener func(session *cherrySession.Session) (next bool, err error)

	RPCHandler func(session *cherrySession.Session, msg *cherryMessage.Message, noCopy bool)

	// process packet listener function
	PacketListener func(agent *Agent, packet *cherryPacket.Packet)

	Options struct {
		Heartbeat        time.Duration                        // heartbeat(sec)
		DataCompression  bool                                 // data compression
		PacketDecoder    cherryPacket.Decoder                 // binary packet decoder
		PacketEncoder    cherryPacket.Encoder                 // binary packet encoder
		Serializer       cherryFacade.ISerializer             // data serializer
		PacketListener   map[cherryPacket.Type]PacketListener // process packet listener function
		RPCHandler       RPCHandler                           // rpc handler
		OnCreateListener []SessionListener                    // on create execute listener function
		OnCloseListener  []SessionListener                    // on close execute listener function
	}

	Agent struct {
		Options
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

func NewAgent(opt Options, session *cherrySession.Session, conn cherryFacade.INetConn) *Agent {
	agent := &Agent{
		Options: opt,
		Session: session,
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

	defer func() {
		if e := recover(); e != nil {
			err = cherryError.ClusterBrokenPipe
		}
	}()

	p := pendingMessage{typ: typ, mid: mid, route: route, payload: v}

	a.chSend <- p

	return nil
}

// Push, implementation for session.NetworkEntity interface
func (a *Agent) Push(route string, v interface{}) error {
	return a.Send(cherryMessage.TypePush, route, 0, v)
}

// RPC, implementation for session.NetworkEntity interface
func (a *Agent) RPC(route string, v interface{}) error {
	if a.Status() == Closed {
		return cherryError.ClusterBrokenPipe
	}

	if a.RPCHandler == nil {
		return cherryError.ClusterRPCHandleNotFound
	}

	data, err := a.Serializer.Marshal(v)
	if err != nil {
		return err
	}

	msg := &cherryMessage.Message{
		Type:  cherryMessage.TypeNotify,
		Route: route,
		Data:  data,
	}

	a.RPCHandler(a.Session, msg, true)

	return nil
}

// Response, implementation for session.NetworkEntity interface
// Response message to session
func (a *Agent) Response(mid uint64, v interface{}) error {
	return a.Send(cherryMessage.TypeResponse, "", mid, v)
}

// Kick
func (a *Agent) Kick(reason string) error {
	pkg, err := a.PacketEncoder.Encode(cherryPacket.Kick, nil)
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

	if cherryProfile.Debug() {
		a.Session.Debugf("session closed.")
	}

	for _, listener := range a.OnCloseListener {
		next, err := listener(a.Session)
		if err != nil {
			cherryLogger.Error(err)
		}

		if next == false {
			break
		}
	}

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
	for _, listener := range a.OnCreateListener {
		next, err := listener(a.Session)
		if err != nil {
			cherryLogger.Error(err)
		}
		if next == false {
			break
		}
	}

	go a.read()
	go a.write()
}

func (a *Agent) read() {
	defer func() {
		a.Close()
	}()

	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			a.Session.Debugf("will be closed immediately. error[%s]", err.Error())
			return
		}

		packets, err := a.PacketDecoder.Decode(msg)
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

func (a *Agent) processPacket(packet *cherryPacket.Packet) {
	listener, found := a.PacketListener[packet.Type]
	if found == false {
		a.Session.Debugf("packet[%s] type not found.", packet)
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
			payload, err := a.Serializer.Marshal(data.payload)
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
			p, err := a.PacketEncoder.Encode(cherryPacket.Data, em)
			if err != nil {
				cherryLogger.Warn(err)
				break
			}
			a.chWrite <- p
		}
	}
}
