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
	"log"
	"net"
	"sync/atomic"
	"time"
)

const (
	agentWriteBacklog = 128
)

const (
	Init = iota
	WaitAck
	Working
	Closed
)

type (
	SessionListener func(session *cherrySession.Session) (next bool, err error)

	RpcHandler func(session *cherrySession.Session, msg *cherryMessage.Message, noCopy bool)

	// process packet listener function
	PacketListener func(agent *Agent, packet *cherryPacket.Packet)

	Options struct {
		Heartbeat        int                      // heartbeat(sec)
		DataCompression  bool                     // data compression
		PacketDecoder    cherryPacket.Decoder     // binary packet decoder
		PacketEncoder    cherryPacket.Encoder     // binary packet encoder
		Serializer       cherryFacade.ISerializer // data serializer
		PacketListener   PacketListener           // process packet listener function
		OnCreateListener []SessionListener        // on create execute listener function
		OnCloseListener  []SessionListener        // on close execute listener function
	}

	Agent struct {
		Options
		Session    *cherrySession.Session // session
		Conn       cherryFacade.Conn      // low-level conn fd
		RpcHandler RpcHandler             // rpc client invoke handler
		state      int32                  // current session state
		chDie      chan bool              // wait for close
		chSend     chan pendingMessage    // push message queue
		chWrite    chan []byte            // push bytes queue
		lastAt     int64                  // last heartbeat unix time stamp
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

func NewAgent(opt Options, session *cherrySession.Session, conn cherryFacade.Conn, rpcHandler RpcHandler) *Agent {
	agent := &Agent{
		Options:    opt,
		Session:    session,
		Conn:       conn,
		RpcHandler: rpcHandler,
		state:      Init,
		chDie:      make(chan bool),
		chSend:     make(chan pendingMessage, agentWriteBacklog),
		chWrite:    make(chan []byte, agentWriteBacklog),
		lastAt:     time.Now().Unix(),
	}

	return agent
}

func (a *Agent) Status() int32 {
	return atomic.LoadInt32(&a.state)
}

func (a *Agent) SetStatus(state int32) {
	atomic.StoreInt32(&a.state, state)
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

	data, err := a.Serializer.Marshal(v)
	if err != nil {
		return err
	}

	msg := &cherryMessage.Message{
		Type:  cherryMessage.TypeNotify,
		Route: route,
		Data:  data,
	}

	a.RpcHandler(a.Session, msg, true)
	return nil
}

// ResponseMid, implementation for session.NetworkEntity interface
// Response message to session
func (a *Agent) ResponseMid(mid uint64, v interface{}) error {
	return a.Send(cherryMessage.TypeResponse, "", mid, v)
}

// Close closes the Agent, clean inner state and close low-level connection.
func (a *Agent) Close() {
	if a.Status() == Closed {
		return
	}

	a.SetStatus(Closed)

	if cherryProfile.Debug() {
		cherryLogger.Debugf("session closed. [%s]", a.Session)
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

	// prevent closing closed channel
	select {
	case <-a.chDie:
		// expect
	default:
		close(a.chDie)
	}

	if err := a.Conn.Close(); err != nil {
		cherryLogger.Errorf("session close error. session:%s, error:%v", a.Session, err)
	}
}

// RemoteAddr, implementation for session.NetworkEntity interface
// returns the remote network address.
func (a *Agent) RemoteAddr() net.Addr {
	return a.Conn.RemoteAddr()
}

// String, implementation for Stringer interface
func (a *Agent) String() string {
	return fmt.Sprintf("Remote=%s, LastTime=%d",
		a.Conn.RemoteAddr().String(),
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
		a.chDie <- true
	}()

	for {
		msg, err := a.Conn.GetNextMessage()
		if err != nil {
			cherryLogger.Debugf("error: %s, session will be closed immediately.", err.Error())
			return
		}

		packets, err := a.PacketDecoder.Decode(msg)
		if err != nil {
			cherryLogger.Warn(err)
			continue
		}

		if len(packets) < 1 {
			continue
		}

		for _, packet := range packets {
			a.PacketListener(a, packet)
		}
	}
}

func (a *Agent) write() {
	// clean func
	defer func() {
		close(a.chSend)
		close(a.chWrite)
		a.Close()
	}()

	for {
		select {
		case bytes := <-a.chWrite:
			// close Agent while low-level conn broken
			_, err := a.Conn.Write(bytes)
			if err != nil {
				cherryLogger.Error(err)
				return
			}

		case data := <-a.chSend:
			payload, err := a.Serializer.Marshal(data.payload)
			if err != nil {
				cherryLogger.Debugf("message serializer error. %s", data.String())
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
				log.Println(err)
				break
			}
			a.chWrite <- p

		case <-a.chDie: // Agent closed signal
			return

			//case <-env.Die: // application quit
			//	return
			//}
		}
	}
}
