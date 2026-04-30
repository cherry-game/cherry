package pomelo

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	ctime "github.com/cherry-game/cherry/extend/time"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	pomeloMessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	pomeloPacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
	cproto "github.com/cherry-game/cherry/net/proto"
	"go.uber.org/zap/zapcore"
)

const (
	AgentInit    int32 = 0
	AgentWaitAck int32 = 1
	AgentWorking int32 = 2
	AgentClosed  int32 = 3
)

type (
	Agent struct {
		cfacade.IApplication                      // app
		conn                 net.Conn             // low-level conn fd
		state                atomic.Int32         // current agent state
		session              *cproto.Session      // session
		chDie                chan struct{}        // signal writeChan to exit
		chPending            chan *pendingMessage // push message queue
		chWrite              chan []byte          // push bytes queue
		lastAt               atomic.Int64         // last heartbeat unix time stamp
		onCloseFunc          []OnCloseFunc        // on close agent
		closeOnce            sync.Once            // ensure Close logic runs once
	}

	pendingMessage struct {
		typ     pomeloMessage.Type // message type
		route   string             // message route(push)
		mid     uint               // response message id(response)
		payload interface{}        // payload
		err     bool               // if it's an error
	}
	OnCloseFunc func(*Agent)
)

func NewAgent(app cfacade.IApplication, conn net.Conn, session *cproto.Session) *Agent {
	agent := &Agent{
		IApplication: app,
		conn:         conn,
		session:      session,
		chDie:        make(chan struct{}),
		chPending:    make(chan *pendingMessage, cmd.writeBacklog),
		chWrite:      make(chan []byte, cmd.writeBacklog),
	}

	agent.state.Store(AgentInit)
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
	return a.state.Load()
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

func (a *Agent) IsBind() bool {
	return a.session.Uid > 0
}

func (a *Agent) Unbind() {
	Unbind(a.SID())
}

func (a *Agent) SetLastAt() {
	a.lastAt.Store(ctime.Now().ToSecond())
}

func (a *Agent) SendRaw(bytes []byte) {
	if a.state.Load() == AgentClosed {
		return
	}
	select {
	case a.chWrite <- bytes:
	default:
	}
}

func (a *Agent) SendPacket(typ pomeloPacket.Type, data []byte) {
	pkg, err := pomeloPacket.Encode(typ, data)
	if err != nil {
		clog.Warn(err)
		return
	}
	a.SendRaw(pkg)
}

// Close sets state to AgentClosed, signals writeChan to exit, runs cleanup
// callbacks, unbinds the session, closes the connection, and drains channels.
// Safe to call multiple times — only the first call takes effect.
func (a *Agent) Close() {
	a.closeOnce.Do(func() {
		a.state.Store(AgentClosed)
		close(a.chDie)

		cutils.Try(func() {
			for _, fn := range a.onCloseFunc {
				fn(a)
			}
		}, func(errString string) {
			clog.Warnf("[sid = %s,uid = %d] Agent on close func error. error = %s",
				a.SID(),
				a.UID(),
				errString)
		})

		a.Unbind()

		closeErr := a.conn.Close()

		if clog.PrintLevel(zapcore.InfoLevel) {
			clog.Infof("[sid = %s,uid = %d] Agent closed. [count = %d, ip = %s, error = %v]",
				a.SID(),
				a.UID(),
				Count(),
				a.RemoteAddr(),
				closeErr,
			)
		}

		a.drainChannels()
	})
}

func (a *Agent) drainChannels() {
	for {
		select {
		case <-a.chPending:
		case <-a.chWrite:
		default:
			return
		}
	}
}

func (a *Agent) Run() {
	go a.writeChan()
	go a.readChan()
}

func (a *Agent) readChan() {
	defer func() {
		// ensure Close is called when readChan exits,
		// no-op if already closed by writeChan or other caller
		a.Close()
	}()

	for {
		packets, isBreak, err := pomeloPacket.Read(a.conn)
		if isBreak || err != nil {
			if clog.PrintLevel(zapcore.InfoLevel) {
				clog.Infof("[sid = %s,uid = %d] Agent read chan exit. [isBreak = %v, error = %v]",
					a.SID(),
					a.UID(),
					isBreak,
					err,
				)
			}
			return
		}

		if len(packets) < 1 {
			continue
		}

		for _, packet := range packets {
			a.processPacket(packet)
		}
	}
}

func (a *Agent) writeChan() {
	ticker := time.NewTicker(cmd.heartbeatTime)
	defer func() {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[sid = %s,uid = %d] Agent write chan exit.", a.SID(), a.UID())
		}
		ticker.Stop()
		// ensure Close is called when writeChan exits,
		// no-op if already closed by readChan or other caller
		a.Close()
	}()

	var lastAt, deadline int64

	for {
		select {
		case <-a.chDie:
			return
		case <-ticker.C:
			lastAt = a.lastAt.Load()
			deadline = time.Now().Add(-cmd.heartbeatTime).Unix()
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
				clog.Warnf("[sid = %s,uid = %d] Write bytes err=%v", a.SID(), a.UID(), err)
				return
			}
		}
	}
}

func (a *Agent) write(bytes []byte) error {
	if a.state.Load() == AgentClosed {
		if clog.PrintLevel(zapcore.InfoLevel) {
			clog.Infof("[sid = %s,uid = %d] Write bytes failed because the connection is closed!", a.SID(), a.UID())
		}
		return nil
	}
	_, err := a.conn.Write(bytes)
	return err
}

func (a *Agent) processPacket(packet *pomeloPacket.Packet) {
	process, found := cmd.onPacketFuncMap[packet.Type()]
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Packet type not found, close connect! [packet = %+v]",
				a.SID(),
				a.UID(),
				packet,
			)
		}
		a.Close()
		return
	}
	process(a, packet)
	a.SetLastAt()
}

func (a *Agent) RemoteAddr() string {
	if a.session != nil {
		return a.session.Ip
	}
	return ""
}

func (p *pendingMessage) String() string {
	return fmt.Sprintf("typ = %d, route = %s, mid = %d, payload = %v", p.typ, p.route, p.mid, p.payload)
}

func (a *Agent) processPending(data *pendingMessage) {
	payload, err := a.Serializer().Marshal(data.payload)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Payload marshal error. [data = %s]",
			a.SID(),
			a.UID(),
			data.String(),
		)
		return
	}

	m := &pomeloMessage.Message{
		Type:  data.typ,
		ID:    data.mid,
		Route: data.route,
		Data:  payload,
		Error: data.err,
	}

	em, err := pomeloMessage.Encode(m)
	if err != nil {
		clog.Warn(err)
		return
	}

	a.SendPacket(pomeloPacket.Data, em)
}

func (a *Agent) sendPending(typ pomeloMessage.Type, route string, mid uint32, v interface{}, isError bool) {
	if a.state.Load() == AgentClosed {
		clog.Warnf("[sid = %s,uid = %d] Session is closed. [typ = %v, route = %s, mid = %d, val = %+v, err = %v]",
			a.SID(),
			a.UID(),
			typ,
			route,
			mid,
			v,
			isError,
		)
		return
	}

	pending := &pendingMessage{
		typ:     typ,
		mid:     uint(mid),
		route:   route,
		payload: v,
		err:     isError,
	}

	select {
	case a.chPending <- pending:
	default:
		clog.Warnf("[sid = %s,uid = %d] send buffer exceed. [typ = %v, route = %s, mid = %d, val = %+v, err = %v]",
			a.SID(),
			a.UID(),
			typ,
			route,
			mid,
			v,
			isError,
		)
	}
}

func (a *Agent) Response(session *cproto.Session, v interface{}, isError ...bool) {
	a.ResponseMID(session.GetMID(), v, isError...)
}

func (a *Agent) ResponseCode(session *cproto.Session, statusCode int32, isError ...bool) {
	rsp := &cproto.Response{
		Code: statusCode,
	}
	a.ResponseMID(session.GetMID(), rsp, isError...)
}

func (a *Agent) ResponseMID(mid uint32, v interface{}, isError ...bool) {
	isErr := false
	if len(isError) > 0 {
		isErr = isError[0]
	}

	a.sendPending(pomeloMessage.Response, "", mid, v, isErr)
	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Response ok. [mid = %d, isError = %v]",
			a.SID(),
			a.UID(),
			mid,
			isErr,
		)
	}
}

func (a *Agent) Push(route string, val interface{}) {
	a.sendPending(pomeloMessage.Push, route, 0, val, false)

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Push ok. [route = %s]",
			a.SID(),
			a.UID(),
			route,
		)
	}
}

func (a *Agent) Kick(reason interface{}, closed bool) {
	bytes, err := a.Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Kick marshal fail. [reason = {%+v}, err = %s]",
			a.SID(),
			a.UID(),
			reason,
			err,
		)
		if closed {
			a.Close()
		}
		return
	}

	pkg, err := pomeloPacket.Encode(pomeloPacket.Kick, bytes)
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
	a.SendRaw(pkg)

	if closed {
		a.Close()
	}
}

func (a *Agent) AddOnClose(fn OnCloseFunc) {
	if a.state.Load() != AgentInit {
		clog.Warnf("[sid = %s] AddOnClose failed: agent is not in Init state. [state = %d]",
			a.SID(),
			a.State(),
		)
		return
	}
	if fn != nil {
		a.onCloseFunc = append(a.onCloseFunc, fn)
	}
}
