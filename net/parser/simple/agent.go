package simple

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	cnet "github.com/cherry-game/cherry/extend/net"
	ctime "github.com/cherry-game/cherry/extend/time"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	"go.uber.org/zap/zapcore"
)

// Agent lifecycle states.
const (
	AgentInit   int32 = 0 // initial state
	AgentClosed int32 = 4 // closed, no further processing
)

type (
	// Agent represents a single client connection. All conn.Write and teardown
	// operations happen exclusively in writeChan's goroutine.
	Agent struct {
		cfacade.IApplication
		conn        net.Conn
		state       atomic.Int32
		session     *cproto.Session
		chDie       chan struct{}
		chPending   chan *pendingMessage
		chWrite     chan []byte
		chKick      chan []byte
		lastAt      atomic.Int64
		onCloseFunc []OnCloseFunc
	}

	pendingMessage struct {
		mid     uint32
		payload interface{}
	}

	OnCloseFunc func(*Agent)
)

// NewAgent creates an agent for a new connection.
func NewAgent(app cfacade.IApplication, conn net.Conn, session *cproto.Session) *Agent {
	agent := &Agent{
		IApplication: app,
		conn:         conn,
		session:      session,
		chDie:        make(chan struct{}),
		chPending:    make(chan *pendingMessage, writeBacklog),
		chWrite:      make(chan []byte, writeBacklog),
		chKick:       make(chan []byte, 1),
	}

	agent.state.Store(AgentInit)
	agent.SetLastAt()

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Agent create. [count = %d, ip = %s]",
			agent.SID(), agent.UID(), Count(), agent.RemoteAddr())
	}

	return agent
}

func (a *Agent) State() int32 {
	return a.state.Load()
}

func (a *Agent) IsClosed() bool {
	return a.state.Load() == AgentClosed
}

func (a *Agent) SetState(state int32) bool {
	oldValue := a.state.Swap(state)
	return oldValue != state
}

func (a *Agent) Session() *cproto.Session {
	return a.session
}

func (a *Agent) UID() cfacade.UID {
	return a.session.Uid
}

func (a *Agent) SID() cfacade.SID {
	return a.session.Sid
}

func (a *Agent) Bind(uid cfacade.UID) (*Agent, error) {
	return Bind(a.SID(), uid)
}

func (a *Agent) Unbind() {
	Unbind(a.SID())
}

func (a *Agent) SetLastAt() {
	a.lastAt.Store(ctime.Now().ToSecond())
}

func (a *Agent) SendRaw(bytes []byte) {
	if a.IsClosed() {
		return
	}
	select {
	case a.chWrite <- bytes:
	default:
	}
}

func (a *Agent) Response(mid uint32, v interface{}) {
	a.sendPending(mid, v)
	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Response ok. [mid = %d, val = %+v]",
			a.SID(), a.UID(), mid, v)
	}
}

func (a *Agent) SendKick(pkg []byte) {
	if a.IsClosed() {
		return
	}
	select {
	case a.chKick <- pkg:
	default:
		clog.Warnf("[sid = %s,uid = %d] Kick buffer full, closing without kick packet.", a.SID(), a.UID())
		a.Close()
	}
}

func (a *Agent) Kick(mid uint32, reason interface{}, closed bool) {
	bytes, err := a.Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Kick marshal fail. [closed = %v, reason = {%+v}, err = %s]",
			a.SID(), a.UID(), closed, reason, err)
		if closed {
			a.Close()
		}
		return
	}

	pkg, err := pack(mid, bytes)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Kick packet encode error. [closed = %v, reason = %+v, err = %s]",
			a.SID(), a.UID(), closed, reason, err)
		if closed {
			a.Close()
		}
		return
	}

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Kick ok. [closed = %v, reason = %+v]",
			a.SID(), a.UID(), closed, reason)
	}

	if closed {
		a.SendKick(pkg)
	} else {
		a.SendRaw(pkg)
	}
}

func (a *Agent) Close() {
	if a.SetState(AgentClosed) {
		select {
		case <-a.chDie:
		default:
			close(a.chDie)
		}
		a.conn.SetWriteDeadline(time.Now().Add(-1))
	}
}

func (a *Agent) closeClean() {
	for _, fn := range a.onCloseFunc {
		cutils.Try(func() { fn(a) }, func(errString string) {
			clog.Warnf("[sid = %s,uid = %d] onCloseFunc error = %s",
				a.SID(), a.UID(), errString)
		})
	}

	a.Unbind()

	if err := a.conn.Close(); err != nil {
		clog.Debugf("[sid = %s,uid = %d] Agent connect closed. [error = %s]",
			a.SID(), a.UID(), err)
	}

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Agent closed. [count = %d, ip = %s]",
			a.SID(), a.UID(), Count(), a.RemoteAddr())
	}
}

func (a *Agent) AddOnClose(fn OnCloseFunc) {
	if a.state.Load() != AgentInit {
		clog.Warnf("[sid = %s,uid = %d] AddOnClose failed: agent is not in Init state. [state = %d]",
			a.SID(), a.UID(), a.State())
		return
	}
	if fn != nil {
		a.onCloseFunc = append(a.onCloseFunc, fn)
	}
}

func (a *Agent) Run() {
	go a.writeChan()
	go a.readChan()
}

func (a *Agent) readChan() {
	defer func() {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[sid = %s,uid = %d] Agent read chan exit.", a.SID(), a.UID())
		}
		a.Close()
	}()

	for {
		if a.IsClosed() {
			return
		}

		msg, isBreak, err := ReadMessage(a.conn)
		if isBreak || err != nil {
			return
		}
		a.processPacket(&msg)
	}
}

func (a *Agent) writeChan() {
	ticker := time.NewTicker(heartbeatTime)
	defer func() {
		ticker.Stop()
		a.Close()
		a.closeClean()
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[sid = %s,uid = %d] Agent write chan exit.", a.SID(), a.UID())
		}
	}()

	var lastAt, deadline int64

	for {
		select {
		case <-a.chDie:
			return
		case bytes := <-a.chKick:
			if err := a.write(bytes); err != nil {
				clog.Warnf("[sid = %s,uid = %d] Kick write error. [error = %v]", a.SID(), a.UID(), err)
			}
			return
		case <-ticker.C:
			lastAt = a.lastAt.Load()
			deadline = time.Now().Add(-heartbeatTime).Unix()
			if lastAt < deadline {
				if clog.PrintLevel(zapcore.DebugLevel) {
					clog.Debugf("[sid = %s,uid = %d] Check heartbeat timeout.", a.SID(), a.UID())
				}
				return
			}
		case pending := <-a.chPending:
			a.processPending(pending)
		case bytes := <-a.chWrite:
			if err := a.write(bytes); err != nil {
				clog.Warnf("[sid = %s,uid = %d] Write bytes error. [error = %v]", a.SID(), a.UID(), err)
				return
			}
		}
	}
}

func (a *Agent) write(bytes []byte) error {
	if a.IsClosed() {
		if clog.PrintLevel(zapcore.InfoLevel) {
			clog.Infof("[sid = %s,uid = %d] Write bytes failed because the connection is closed!", a.SID(), a.UID())
		}
		return nil
	}
	_, err := a.conn.Write(bytes)
	return err
}

func (a *Agent) processPacket(msg *Message) {
	nodeRoute, found := GetNodeRoute(msg.MID)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Route not found, close connect! [message = %+v]",
				a.SID(), a.UID(), msg)
		}
		a.Close()
		return
	}
	onDataRouteFunc(a, msg, nodeRoute)
	a.SetLastAt()
}

func (a *Agent) processPending(pending *pendingMessage) {
	data, err := a.Serializer().Marshal(pending.payload)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Payload marshal error. [data = %s]",
			a.SID(), a.UID(), pending.String())
		return
	}

	pkg, err := pack(pending.mid, data)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Pack error. [error = %v]", a.SID(), a.UID(), err)
		return
	}

	a.SendRaw(pkg)
}

func (a *Agent) sendPending(mid uint32, payload interface{}) {
	if a.IsClosed() {
		clog.Warnf("[sid = %s,uid = %d] Session is closed. [mid = %d, payload = %+v]",
			a.SID(), a.UID(), mid, payload)
		return
	}

	pending := &pendingMessage{
		mid:     mid,
		payload: payload,
	}

	select {
	case a.chPending <- pending:
	default:
		clog.Warnf("[sid = %s,uid = %d] send buffer exceed. [mid = %d, payload = %+v]",
			a.SID(), a.UID(), mid, payload)
	}
}

func (a *Agent) RemoteAddr() string {
	if a.conn != nil {
		return cnet.GetIPV4(a.conn.RemoteAddr())
	}
	return ""
}

func (p *pendingMessage) String() string {
	return fmt.Sprintf("mid = %d, payload = %v", p.mid, p.payload)
}
