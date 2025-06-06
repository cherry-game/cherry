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

const (
	AgentInit   int32 = 0
	AgentClosed int32 = 3
)

type (
	Agent struct {
		cfacade.IApplication                      // app
		conn                 net.Conn             // low-level conn fd
		state                int32                // current agent state
		session              *cproto.Session      // session
		chDie                chan struct{}        // wait for close
		chPending            chan *pendingMessage // push message queue
		chWrite              chan []byte          // push bytes queue
		lastAt               int64                // last heartbeat unix time stamp
		onCloseFunc          []OnCloseFunc        // on close agent
	}

	pendingMessage struct {
		mid     uint32
		payload interface{}
	}

	OnCloseFunc func(*Agent)
)

func NewAgent(app cfacade.IApplication, conn net.Conn, session *cproto.Session) Agent {
	agent := Agent{
		IApplication: app,
		conn:         conn,
		state:        AgentInit,
		session:      session,
		chDie:        make(chan struct{}),
		chPending:    make(chan *pendingMessage, writeBacklog),
		chWrite:      make(chan []byte, writeBacklog),
		lastAt:       0,
		onCloseFunc:  nil,
	}

	agent.session.Ip = agent.RemoteAddr()
	agent.SetLastAt()

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Agent create. [count = %d, ip = %s]",
			agent.SID(),
			agent.UID(),
			Count(),
			agent.RemoteAddr(),
		)
	}

	return agent
}

func (a *Agent) State() int32 {
	return a.state
}

func (a *Agent) SetState(state int32) bool {
	oldValue := atomic.SwapInt32(&a.state, state)
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

func (a *Agent) Bind(uid cfacade.UID) error {
	return BindUID(a.SID(), uid)
}

func (a *Agent) Unbind() {
	Unbind(a.SID())
}

func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, ctime.Now().ToSecond())
}

func (a *Agent) SendRaw(bytes []byte) {
	a.chWrite <- bytes
}

func (a *Agent) Close() {
	if a.SetState(AgentClosed) {
		select {
		case <-a.chDie:
		default:
			close(a.chDie)
		}
	}
}

func (a *Agent) Run() {
	go a.writeChan()
	go a.readChan()
}

func (a *Agent) readChan() {
	defer func() {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[sid = %s,uid = %d] Agent read chan exit.",
				a.SID(),
				a.UID(),
			)
		}

		a.Close()
	}()

	for {
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
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[sid = %s,uid = %d] Agent write chan exit.", a.SID(), a.UID())
		}

		ticker.Stop()
		a.closeProcess()
		a.Close()
	}()

	for {
		select {
		case <-a.chDie:
			{
				return
			}
		case <-ticker.C:
			{
				deadline := ctime.Now().Add(-heartbeatTime).Unix()
				if a.lastAt < deadline {
					if clog.PrintLevel(zapcore.DebugLevel) {
						clog.Debugf("[sid = %s,uid = %d] Check heartbeat timeout.", a.SID(), a.UID())
					}
					return
				}
			}
		case pending := <-a.chPending:
			{
				a.processPending(pending)
			}
		case bytes := <-a.chWrite:
			{
				a.write(bytes)
			}
		}
	}
}

func (a *Agent) closeProcess() {
	cutils.Try(func() {
		for _, fn := range a.onCloseFunc {
			fn(a)
		}
	}, func(errString string) {
		clog.Warn(errString)
	})

	a.Unbind()

	if err := a.conn.Close(); err != nil {
		clog.Debugf("[sid = %s,uid = %d] Agent connect closed. [error = %s]",
			a.SID(),
			a.UID(),
			err,
		)
	}

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Agent closed. [count = %d, ip = %s]",
			a.SID(),
			a.UID(),
			Count(),
			a.RemoteAddr(),
		)
	}

	close(a.chPending)
	close(a.chWrite)
}

func (a *Agent) write(bytes []byte) {
	_, err := a.conn.Write(bytes)
	if err != nil {
		clog.Warn(err)
	}
}

func (a *Agent) processPacket(msg *Message) {
	nodeRoute, found := GetNodeRoute(msg.MID)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Route not found, close connect! [message = %+v]",
				a.SID(),
				a.UID(),
				msg,
			)
		}
		a.Close()
		return
	}

	onDataRouteFunc(a, msg, nodeRoute)

	// update last time
	a.SetLastAt()
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

func (a *Agent) processPending(pending *pendingMessage) {
	data, err := a.Serializer().Marshal(pending.payload)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Payload marshal error. [data = %s]",
			a.SID(),
			a.UID(),
			pending.String(),
		)
		return
	}

	// encode packet
	pkg, err := pack(pending.mid, data)
	if err != nil {
		clog.Warn(err)
		return
	}

	a.SendRaw(pkg)
}

func (a *Agent) sendPending(mid uint32, payload interface{}) {
	if a.state == AgentClosed {
		clog.Warnf("[sid = %s,uid = %d] Session is closed. [mid = %d, payload = %+v]",
			a.SID(),
			a.UID(),
			mid,
			payload,
		)
		return
	}

	if len(a.chPending) >= writeBacklog {
		clog.Warnf("[sid = %s,uid = %d] send buffer exceed. [mid = %d, payload = %+v]",
			a.SID(),
			a.UID(),
			mid,
			payload,
		)
		return
	}

	pending := &pendingMessage{
		mid:     mid,
		payload: payload,
	}

	a.chPending <- pending
}

func (a *Agent) Response(mid uint32, v interface{}) {
	a.sendPending(mid, v)
	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Response ok. [mid = %d, val = %+v]",
			a.SID(),
			a.UID(),
			mid,
			v,
		)
	}
}

func (a *Agent) AddOnClose(fn OnCloseFunc) {
	if fn != nil {
		a.onCloseFunc = append(a.onCloseFunc, fn)
	}
}

func (a *Agent) Kick(mid uint32, reason interface{}, closed bool) {
	bytes, err := a.Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Kick marshal fail. [reason = {%+v}, err = %s]",
			a.SID(),
			a.UID(),
			reason,
			err,
		)
	}
	// encode packet
	pkg, err := pack(mid, bytes)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Kick packet encode error.[reason = %+v, err = %s]",
			a.SID(),
			a.UID(),
			reason,
			err,
		)
		return
	}

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Kick ok. [reason = %+v, closed = %v]",
			a.SID(),
			a.UID(),
			reason,
			closed,
		)
	}

	// 不进入pending chan，直接踢了
	a.write(pkg)

	if closed {
		a.Close()
	}
}
