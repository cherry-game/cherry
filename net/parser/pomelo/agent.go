package pomelo

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	cnet "github.com/cherry-game/cherry/extend/net"
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
		state                int32                // current agent state
		session              *cproto.Session      // session
		chDie                chan struct{}        // wait for close
		chPending            chan *pendingMessage // push message queue
		chWrite              chan []byte          // push bytes queue
		lastAt               int64                // last heartbeat unix time stamp
		onCloseFunc          []OnCloseFunc        // on close agent
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

func NewAgent(app cfacade.IApplication, conn net.Conn, session *cproto.Session) Agent {
	agent := Agent{
		IApplication: app,
		conn:         conn,
		state:        AgentInit,
		session:      session,
		chDie:        make(chan struct{}),
		chPending:    make(chan *pendingMessage, cmd.writeBacklog),
		chWrite:      make(chan []byte, cmd.writeBacklog),
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

func (a *Agent) IsBind() bool {
	return a.session.Uid > 0
}

func (a *Agent) Unbind() {
	Unbind(a.SID())
}

func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, time.Now().Unix())
}

func (a *Agent) SendRaw(bytes []byte) {
	a.chWrite <- bytes
}

func (a *Agent) SendPacket(typ pomeloPacket.Type, data []byte) {
	pkg, err := pomeloPacket.Encode(typ, data)
	if err != nil {
		clog.Warn(err)
		return
	}
	a.SendRaw(pkg)
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
		packets, isBreak, err := pomeloPacket.Read(a.conn)
		if isBreak || err != nil {
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
		a.closeProcess()
		a.Close()
	}()

	var lastAt, deadline int64

	for {
		select {
		case <-a.chDie:
			{
				return
			}
		case <-ticker.C:
			{
				lastAt = atomic.LoadInt64(&a.lastAt)
				deadline = time.Now().Add(-cmd.heartbeatTime).Unix()
				if lastAt < deadline {
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

	// construct message and encode
	m := &pomeloMessage.Message{
		Type:  data.typ,
		ID:    data.mid,
		Route: data.route,
		Data:  payload,
		Error: data.err,
	}

	// encode message
	em, err := pomeloMessage.Encode(m)
	if err != nil {
		clog.Warn(err)
		return
	}

	// encode packet
	a.SendPacket(pomeloPacket.Data, em)
}

func (a *Agent) sendPending(typ pomeloMessage.Type, route string, mid uint32, v interface{}, isError bool) {
	if a.state == AgentClosed {
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

	if len(a.chPending) >= cmd.writeBacklog {
		clog.Warnf("[sid = %s,uid = %d] send buffer exceed. [typ = %v, route = %s, mid = %d, val = %+v, err = %v]",
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

	a.chPending <- pending
}

func (a *Agent) Response(session *cproto.Session, v interface{}, isError ...bool) {
	a.ResponseMID(session.Mid, v, isError...)
}

func (a *Agent) ResponseCode(session *cproto.Session, statusCode int32, isError ...bool) {
	rsp := &cproto.Response{
		Code: statusCode,
	}
	a.ResponseMID(session.Mid, rsp, isError...)
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

	// 不进入pending chan，直接踢了
	a.write(pkg)

	if closed {
		a.Close()
	}
}

func (a *Agent) AddOnClose(fn OnCloseFunc) {
	if fn != nil {
		a.onCloseFunc = append(a.onCloseFunc, fn)
	}
}
