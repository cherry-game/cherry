package pomelo

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
	pomeloMessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	pomeloPacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
	cproto "github.com/cherry-game/cherry/net/proto"
	"go.uber.org/zap/zapcore"
)

// Agent lifecycle states.
const (
	AgentInit    int32 = 0 // initial state, before handshake
	AgentWaitAck int32 = 1 // waiting for handshake ack
	AgentWorking int32 = 2 // normal operation
	AgentClosed  int32 = 4 // closed, no further processing
)

type (
	// Agent represents a single client connection. All conn.Write and teardown
	// operations happen exclusively in writeChan's goroutine.
	Agent struct {
		cfacade.IApplication                      // app
		conn                 net.Conn             // low-level conn fd
		state                atomic.Int32         // current agent state
		sid                  cfacade.SID          // connection identity, set once at creation
		uid                  atomic.Int64         // bound user id: written by Bind, read from several goroutines
		ip                   string               // connection identity, set once at creation
		agentPath            string               // the parser actor this connection belongs to, set once at creation
		data                 *cproto.SessionData  // the connection's session key/values
		chDie                chan struct{}        // signal writeChan to exit
		chPending            chan *pendingMessage // push message queue
		chWrite              chan []byte          // push bytes queue
		chKick               chan []byte          // kick channel: write bytes then close within writeChan
		lastAt               atomic.Int64         // last heartbeat unix time stamp
		onCloseFunc          []OnCloseFunc        // on close agent
	}

	// pendingMessage is a deferred message waiting to be encoded and sent.
	pendingMessage struct {
		typ     pomeloMessage.Type // message type
		route   string             // message route (push)
		mid     uint               // response message id (response)
		payload interface{}        // payload
		err     bool               // if it's an error
	}

	// OnCloseFunc is a callback invoked during agent teardown. Each callback
	// is individually guarded by Try/catch so one panic never skips the rest.
	OnCloseFunc func(*Agent)
)

// NewAgent creates an agent for a new connection. It is the only way to build a
// usable one: every field but the embedded application is unexported, so an Agent
// literal written elsewhere has no connection, no channels and no session data.
// Binding it into the lookup maps afterwards is the caller's job — see BindSID.
//
// sid and agentPath are connection identity: they are captured here and never
// change, so any goroutine may read them. The session key/values start empty and
// are reached through Data.
func NewAgent(app cfacade.IApplication, conn net.Conn, sid cfacade.SID, agentPath string) *Agent {
	agent := &Agent{
		IApplication: app,
		conn:         conn,
		sid:          sid,
		ip:           cnet.GetIPV4(conn.RemoteAddr()),
		agentPath:    agentPath,
		chDie:        make(chan struct{}),
		chPending:    make(chan *pendingMessage, cmd.writeBacklog),
		chWrite:      make(chan []byte, cmd.writeBacklog),
		chKick:       make(chan []byte, 1),
		data:         &cproto.SessionData{},
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

// State returns the current agent lifecycle state.
func (a *Agent) State() int32 {
	return a.state.Load()
}

// IsClosed returns true if the agent has been closed.
func (a *Agent) IsClosed() bool {
	return a.state.Load() == AgentClosed
}

// SetState atomically swaps the agent state. Returns true if the state changed.
func (a *Agent) SetState(state int32) bool {
	oldValue := a.state.Swap(state)
	return oldValue != state
}

// SID returns the session id.
func (a *Agent) SID() cfacade.SID {
	return a.sid
}

// UID returns the bound user id.
func (a *Agent) UID() cfacade.UID {
	return a.uid.Load()
}

// SetUID binds a user id to this connection.
func (a *Agent) SetUID(uid cfacade.UID) {
	a.uid.Store(uid)
}

// IP returns the client ip address.
func (a *Agent) IP() string {
	return a.ip
}

// AgentPath returns the parser actor path this connection belongs to.
func (a *Agent) AgentPath() string {
	return a.agentPath
}

// Data returns this connection's session key/values. Its methods are safe to
// call from any goroutine.
func (a *Agent) Data() *cproto.SessionData {
	return a.data
}

// GetData reads one session value off this connection — shorthand for Data().Get,
// for the common case of reading a single key.
func (a *Agent) GetData(key string) string {
	return a.data.Get(key)
}

// SetData writes one session value on this connection — shorthand for Data().Set.
func (a *Agent) SetData(key, value string) {
	a.data.Set(key, value)
}

// NewSession builds a session for one message of this connection: a new object
// carrying the connection's identity, a copy of its session data, and mid.
//
// A new object on every call is the point. The message id belongs to the message,
// and this used to be stamped onto the connection's one shared session while the
// connection's actor was reading that same session — the data race that killed
// the process with "concurrent map read and map write".
//
// Pass 0 where a session is needed but no message is being routed, such as when
// syncing the session to another node.
func (a *Agent) NewSession(mid uint) *cproto.Session {
	return &cproto.Session{
		Sid:       a.sid,
		Uid:       a.UID(),
		AgentPath: a.agentPath,
		Ip:        a.ip,
		Data:      a.data.Clone(),
		Mid:       uint32(mid),
	}
}

// Bind associates a uid with this agent's sid. Returns any previously bound
// agent for the same uid (duplicate login).
func (a *Agent) Bind(uid cfacade.UID) (*Agent, error) {
	return Bind(a.SID(), uid)
}

// IsBind returns true if a user has been bound to this session.
func (a *Agent) IsBind() bool {
	return a.uid.Load() > 0
}

// Unbind removes this agent from the global sid and uid maps.
func (a *Agent) Unbind() {
	Unbind(a.SID())
}

// SetLastAt updates the last heartbeat timestamp to now.
func (a *Agent) SetLastAt() {
	a.lastAt.Store(ctime.Now().ToSecond())
}

// SendRaw enqueues raw bytes for writing. Non-blocking; drops data if the
// agent is closed or the write buffer is full.
func (a *Agent) SendRaw(bytes []byte) {
	if a.IsClosed() {
		return
	}

	select {
	case a.chWrite <- bytes:
	default:
	}
}

// SendPacket encodes data with the given packet type and enqueues it.
func (a *Agent) SendPacket(typ pomeloPacket.Type, data []byte) {
	pkg, err := pomeloPacket.Encode(typ, data)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Packet encode error. [error = %v]", a.SID(), a.UID(), err)
		return
	}
	a.SendRaw(pkg)
}

// Response sends a response payload for the given session.
func (a *Agent) Response(session *cproto.Session, v interface{}, isError ...bool) {
	a.ResponseMID(session.GetMID(), v, isError...)
}

// ResponseCode sends a status-code response for the given session.
func (a *Agent) ResponseCode(session *cproto.Session, statusCode int32, isError ...bool) {
	rsp := &cproto.Response{Code: statusCode}
	a.ResponseMID(session.GetMID(), rsp, isError...)
}

// ResponseMID sends a response payload for the given message id.
func (a *Agent) ResponseMID(mid uint32, v interface{}, isError ...bool) {
	isErr := false
	if len(isError) > 0 {
		isErr = isError[0]
	}
	a.sendPending(pomeloMessage.Response, "", mid, v, isErr)
	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Response ok. [mid = %d, isError = %v]",
			a.SID(), a.UID(), mid, isErr)
	}
}

// Push sends a server-push message to the client on the given route.
func (a *Agent) Push(route string, val interface{}) {
	a.sendPending(pomeloMessage.Push, route, 0, val, false)
	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[sid = %s,uid = %d] Push ok. [route = %s]", a.SID(), a.UID(), route)
	}
}

// SendKick enqueues a kick packet via chKick. On success, writeChan writes
// the bytes then tears down. If the channel buffer is full (writeChan is
// stuck on conn.Write), it signals Close to unblock and tear down.
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

// Kick marshals the reason, encodes a kick packet, and sends it.
// If closed is true, the agent is torn down after the kick is written.
func (a *Agent) Kick(reason interface{}, closed bool) {
	bytes, err := a.Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Kick marshal fail. [closed = %v, reason = {%+v}, err = %s]", a.SID(), a.UID(), closed, reason, err)
		if closed {
			a.Close()
		}
		return
	}

	pkg, err := pomeloPacket.Encode(pomeloPacket.Kick, bytes)
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

// Close signals the agent to shut down by setting state to AgentClosed,
// closing chDie, and forcing a write deadline to unblock writeChan if it
// is stuck on conn.Write. Actual teardown (conn.Close, Unbind, callbacks)
// happens in writeChan's defer via closeClean.
// Safe to call multiple times — only the first call takes effect.
func (a *Agent) Close() {
	if a.SetState(AgentClosed) {
		select {
		case <-a.chDie:
		default:
			close(a.chDie)
		}
		// force unblock writeChan if stuck on conn.Write
		a.conn.SetWriteDeadline(time.Now().Add(-1))
	}
}

// closeClean tears down the agent: runs onClose callbacks, unbinds,
// closes the connection, and logs.
func (a *Agent) closeClean() {
	for _, fn := range a.onCloseFunc {
		cutils.Try(func() { fn(a) }, func(errString string) {
			clog.Warnf("[sid = %s,uid = %d] onCloseFunc error = %s",
				a.SID(), a.UID(), errString)
		})
	}

	// remove from sidAgentMap and uidMap
	a.Unbind()

	closeErr := a.conn.Close()
	if clog.PrintLevel(zapcore.InfoLevel) {
		clog.Infof("[sid = %s,uid = %d] Agent closed. [count = %d, ip = %s, error = %v]",
			a.SID(), a.UID(), Count(), a.RemoteAddr(), closeErr)
	}
}

// AddOnClose registers a callback to be invoked during agent teardown.
// Must be called while the agent is in AgentInit state.
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

// Run starts the two goroutines that manage the agent's lifetime:
// readChan reads incoming packets, writeChan writes outgoing data and owns teardown.
func (a *Agent) Run() {
	go a.writeChan()
	go a.readChan()
}

// readChan reads packets from the connection in a loop.
// It exits on read error, connection break, or when the agent is closed.
// On exit, its defer signals Close() to notify writeChan to tear down.
func (a *Agent) readChan() {
	defer func() {
		a.Close()
	}()

	for {
		if a.IsClosed() {
			return
		}

		packets, isBreak, err := pomeloPacket.Read(a.conn)
		if isBreak || err != nil {
			if clog.PrintLevel(zapcore.InfoLevel) {
				clog.Infof("[sid = %s,uid = %d] Agent read chan exit. [isBreak = %v, error = %v]",
					a.SID(), a.UID(), isBreak, err)
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

// writeChan is the sole goroutine that writes to conn and owns the agent
// teardown lifecycle. Its select loop handles:
//
//   - chDie:     external close request → return → defer sets state + closeClean
//   - chKick:    kick packet → write then return → defer tears down
//   - ticker:    heartbeat check → if timeout, return → defer tears down
//   - chPending: pending messages (push/response) → encode and write
//   - chWrite:   raw bytes → write to conn
//
// On any return path, the defer calls Close() (sets state + closes chDie) then closeClean(),
// ensuring conn.Close, Unbind, callbacks, and channel drain all happen in
// this single goroutine.
func (a *Agent) writeChan() {
	ticker := time.NewTicker(cmd.heartbeatTime)
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
				clog.Warnf("[sid = %s,uid = %d] Write bytes error. [error = %v]", a.SID(), a.UID(), err)
				return
			}
		}
	}
}

// write sends bytes to the underlying connection. Returns nil if the agent
// is already closed, allowing the caller to proceed without error.
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

// processPacket dispatches an incoming packet to its registered handler.
// If the packet type is unknown, it signals Close and returns.
func (a *Agent) processPacket(packet *pomeloPacket.Packet) {
	process, found := cmd.onPacketFuncMap[packet.Type()]
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Warnf("[sid = %s,uid = %d] Packet type not found, close connect! [packet = %+v]", a.SID(), a.UID(), packet)
		}
		a.Close()
		return
	}
	process(a, packet)
	a.SetLastAt()
}

// processPending marshals and encodes a pendingMessage, then enqueues the
// resulting packet for writing.
func (a *Agent) processPending(data *pendingMessage) {
	payload, err := a.Serializer().Marshal(data.payload)
	if err != nil {
		clog.Warnf("[sid = %s,uid = %d] Payload marshal error. [data = %s]",
			a.SID(), a.UID(), data.String())
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
		clog.Warnf("[sid = %s,uid = %d] Message encode error. [error = %v]", a.SID(), a.UID(), err)
		return
	}

	a.SendPacket(pomeloPacket.Data, em)
}

// sendPending enqueues a message for deferred encoding and writing.
// Drops the message if the agent is closed or the buffer is full.
func (a *Agent) sendPending(typ pomeloMessage.Type, route string, mid uint32, v interface{}, isError bool) {
	if a.IsClosed() {
		clog.Warnf("[sid = %s,uid = %d] Session is closed. [typ = %v, route = %s, mid = %d, val = %+v, err = %v]",
			a.SID(), a.UID(), typ, route, mid, v, isError)
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
			a.SID(), a.UID(), typ, route, mid, v, isError)
	}
}

// RemoteAddr returns the client IP address.
func (a *Agent) RemoteAddr() string {
	return a.ip
}

// String returns a human-readable representation of the pending message.
func (p *pendingMessage) String() string {
	return fmt.Sprintf("typ = %d, route = %s, mid = %d, payload = %v", p.typ, p.route, p.mid, p.payload)
}
