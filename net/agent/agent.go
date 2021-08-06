package cherryAgent

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryCommand "github.com/cherry-game/cherry/net/command"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

const (
	WriteBacklog = 64
)

type (
	RPCHandler func(session *cherrySession.Session, msg *cherryMessage.Message)

	Options struct {
		Heartbeat  time.Duration                                // heartbeat(sec)
		Commands   map[cherryPacket.Type]cherryCommand.ICommand // commands
		RPCHandler RPCHandler                                   // rpc handler
	}

	Agent struct {
		sync.RWMutex
		*Options
		cherryFacade.IApplication
		Session *cherrySession.Session // session
		conn    cherryFacade.INetConn  // low-level conn fd
		//state   int32                  // current session state
		chDie   chan bool           // wait for close
		chSend  chan pendingMessage // push message queue
		chWrite chan []byte         // push bytes queue
		lastAt  int64               // last heartbeat unix time stamp
	}

	pendingMessage struct {
		typ     cherryMessage.Type // message type
		route   string             // message route(push)
		mid     uint               // response message id(response)
		payload interface{}        // payload
		err     bool               // if its an error
	}
)

func (p *pendingMessage) String() string {
	return fmt.Sprintf("typ = %d, route = %s, mid =%d, payload=%v", p.typ, p.route, p.mid, p.payload)
}

func NewAgent(app cherryFacade.IApplication, conn cherryFacade.INetConn, opts *Options) *Agent {
	agent := &Agent{
		IApplication: app,
		Options:      opts,
		conn:         conn,
		//state:        Init,
		chDie:   make(chan bool),
		chSend:  make(chan pendingMessage, WriteBacklog),
		chWrite: make(chan []byte, WriteBacklog),
		lastAt:  time.Now().Unix(),
	}

	return agent
}

func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, time.Now().Unix())
}

func (a *Agent) SendRaw(bytes []byte) error {
	a.chWrite <- bytes
	return nil
}

func (a *Agent) Send(typ cherryMessage.Type, route string, mid uint, v interface{}, isError bool) (err error) {
	if a.Session.State() == cherrySession.Closed {
		return cherryError.ClusterBrokenPipe
	}

	if len(a.chSend) >= WriteBacklog {
		return cherryError.ClusterBufferExceed
	}

	p := pendingMessage{typ: typ, mid: mid, route: route, payload: v, err: isError}

	a.chSend <- p

	return nil
}

// Push implementation for session.NetworkEntity interface
func (a *Agent) Push(route string, val interface{}) error {
	return a.Send(cherryMessage.Push, route, 0, val, false)
}

// RPC implementation for cherryFacade.INetwork interface
func (a *Agent) RPC(route string, val interface{}) error {
	if a.Session.State() == cherrySession.Closed {
		return cherryError.ClusterBrokenPipe
	}

	if a.RPCHandler == nil {
		return cherryError.ClusterRPCHandleNotFound
	}

	data, err := a.Marshal(val)
	if err != nil {
		return err
	}

	msg := &cherryMessage.Message{
		Type:  cherryMessage.Notify,
		Route: route,
		Data:  data,
	}

	a.RPCHandler(a.Session, msg)

	return nil
}

// Response implementation for session.NetworkEntity interface
func (a *Agent) Response(mid uint, v interface{}, isError ...bool) error {
	err := false
	if len(isError) > 0 {
		err = isError[0]
	}

	return a.Send(cherryMessage.Response, "", mid, v, err)
}

// Kick kick session
func (a *Agent) Kick(reason interface{}) error {
	bytes, err := a.Marshal(reason)
	if err != nil {
		a.Session.Debugf("kick fail. marshal error[%s], reason[%v].", err, reason)
	}

	pkg, err := a.PacketEncode(cherryPacket.Kick, bytes)
	if err != nil {
		return err
	}

	_, err = a.conn.Write(pkg)
	if err != nil {
		cherryLogger.Warn(err)
	}

	a.Session.Debugf("kick session. reason[%v]", reason)

	return nil
}

// Close closes the Agent, clean inner state and close low-level connection.
func (a *Agent) Close() {
	a.Lock()
	defer a.Unlock()

	if a.Session.State() == cherrySession.Closed {
		return
	}
	a.Session.SetState(cherrySession.Closed)

	a.Session.OnCloseProcess()
	a.chDie <- true

	if err := a.conn.Close(); err != nil {
		a.Session.Debugf("session close error[%s]", err)
	}
}

// RemoteAddr implementation for session.NetworkEntity interface
// returns the remote network address.
func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Agent) Run() {
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
			a.Session.Debugf("close read goroutine. error[%s]", err)
			return
		}

		packets, err := a.PacketDecode(msg)
		if err != nil {
			a.Session.Warnf("packet decoder error. error[%s], msg[%s]", err, msg)
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

func (a *Agent) write() {
	ticker := time.NewTicker(a.Heartbeat)
	defer func() {
		a.Session.Debugf("close write goroutine.")

		ticker.Stop()
		a.Close()

		close(a.chSend)
		close(a.chWrite)
		//close(a.chDie)
	}()

	for {
		select {
		case <-a.chDie:
			return

		case <-ticker.C:
			deadline := time.Now().Add(-a.Heartbeat).Unix()
			if a.lastAt < deadline {
				a.Session.Debug("check heartbeat timeout.")
				return
			}

		case bytes := <-a.chWrite:
			_, err := a.conn.Write(bytes)
			if err != nil {
				cherryLogger.Warn(err)
				return
			}

		case data := <-a.chSend:
			payload, err := a.Marshal(data.payload)
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
				Error: data.err,
			}

			// encode message
			em, err := cherryMessage.Encode(m)
			if err != nil {
				cherryLogger.Warn(err)
				break
			}

			// encode packet
			p, err := a.PacketEncode(cherryPacket.Data, em)
			if err != nil {
				cherryLogger.Warn(err)
				break
			}
			a.chWrite <- p
		}
	}
}

func (a *Agent) processPacket(packet cherryFacade.IPacket) {
	cmd, found := a.Commands[packet.Type()]
	if found == false {
		a.Session.Debugf("packet[%s] type not found.", packet)
		return
	}

	if a.Session == nil {
		cherryLogger.Warnf("session is nil.")
		return
	}

	cmd.Do(a.Session, packet)

	// update last time
	a.SetLastAt()
}
