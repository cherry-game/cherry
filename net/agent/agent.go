package cherryAgent

import (
	"fmt"
	ccode "github.com/cherry-game/cherry/code"
	cnet "github.com/cherry-game/cherry/extend/net"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	ccmd "github.com/cherry-game/cherry/net/command"
	cmsg "github.com/cherry-game/cherry/net/message"
	cpkg "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
	"go.uber.org/zap/zapcore"
	"sync"
	"sync/atomic"
	"time"
)

const (
	WriteBacklog = 64
)

type (
	Options struct {
		Heartbeat time.Duration               // heartbeat(sec)
		Commands  map[cpkg.Type]ccmd.ICommand // commands
	}

	Agent struct {
		sync.RWMutex
		*Options
		cfacade.IApplication
		session *csession.Session    // session
		conn    cfacade.INetConn     // low-level conn fd
		chDie   chan bool            // wait for close
		chSend  chan *pendingMessage // push message queue
		chWrite chan []byte          // push bytes queue
		lastAt  int64                // last heartbeat unix time stamp
	}

	pendingMessage struct {
		typ     cmsg.Type   // message type
		route   string      // message route(push)
		mid     uint        // response message id(response)
		payload interface{} // payload
		err     bool        // if its an error
	}
)

func (p *pendingMessage) String() string {
	return fmt.Sprintf("typ = %d, route = %s, mid = %d, payload = %v", p.typ, p.route, p.mid, p.payload)
}

func NewAgent(app cfacade.IApplication, conn cfacade.INetConn, opts *Options) *Agent {
	agent := &Agent{
		IApplication: app,
		Options:      opts,
		conn:         conn,
		chDie:        make(chan bool),
		chSend:       make(chan *pendingMessage, WriteBacklog),
		chWrite:      make(chan []byte, WriteBacklog),
	}

	return agent
}

func (a *Agent) SetSession(session *csession.Session) {
	a.session = session
}

func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, time.Now().Unix())
}

func (a *Agent) SendRaw(bytes []byte) {
	a.chWrite <- bytes
}

func (a *Agent) RPC(nodeId string, route string, req, _ interface{}) int32 {
	clog.Errorf("[RPC] cluster no implement. [nodeId = %s, route = %s, req = {%+v}]", nodeId, route, req)
	return ccode.OK
}

func (a *Agent) Response(mid uint, v interface{}, isError ...bool) {
	isErr := false
	if len(isError) > 0 {
		isErr = isError[0]
	}

	a.send(cmsg.Response, "", mid, v, isErr)
	if clog.PrintLevel(zapcore.DebugLevel) {
		a.session.Debugf("[Response] ok. [mid = %d, isError = %v, val = %+v]", mid, isErr, v)
	}
}

func (a *Agent) Push(route string, val interface{}) {
	a.send(cmsg.Push, route, 0, val, false)

	if clog.PrintLevel(zapcore.DebugLevel) {
		a.session.Debugf("[Push] ok. [route = %s, val = %+v]", route, val)
	}
}

func (a *Agent) Kick(reason interface{}) {
	bytes, err := a.Marshal(reason)
	if err != nil {
		a.session.Warnf("[Kick] marshal fail. [reason = {%+v}, err = %s].", reason, err)
	}

	pkg, err := a.PacketEncode(cpkg.Kick, bytes)
	if err != nil {
		a.session.Warnf("[Kick] packet encode error.[reason = {%+v}, err = %s].", reason, err)
		return
	}

	_, err = a.conn.Write(pkg)
	if err != nil {
		clog.Warn(err)
	}

	if clog.PrintLevel(zapcore.DebugLevel) {
		a.session.Debugf("[Kick] ok. [reason = {%+v}]", reason)
	}
}

func (a *Agent) RemoteAddr() string {
	if a.conn != nil {
		return cnet.GetIPV4(a.conn.RemoteAddr())
	}

	return ""
}

func (a *Agent) Close() {
	a.Lock()
	defer a.Unlock()

	if a.session.State() == csession.Closed {
		return
	}

	a.session.SetState(csession.Closed)
	a.session.OnCloseListener()

	a.chDie <- true

	if err := a.conn.Close(); err != nil {
		a.session.Debugf("session close. [error = %s]", err)
	}
}

func (a *Agent) send(typ cmsg.Type, route string, mid uint, v interface{}, isError bool) {
	if a.session.State() == csession.Closed {
		a.session.Warnf("[Send] session is closed. [typ = %v, route = %s, mid = %d, val = %+v, isErr = %v",
			typ,
			route,
			mid,
			v,
			isError,
		)
		return
	}

	if len(a.chSend) >= WriteBacklog {
		a.session.Warnf("[Send] send buffer exceed. [typ = %v, route = %s, mid = %d, val = %+v, isErr = %v",
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
		mid:     mid,
		route:   route,
		payload: v,
		err:     isError,
	}

	a.chSend <- pending
}

func (a *Agent) Run() {
	if a.session == nil {
		clog.Error("session is nil. run fail.")
		return
	}

	go a.readChan()
	go a.writeChan()
}

func (a *Agent) readChan() {
	defer func() {
		a.Close()
	}()

	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			return
		}

		packets, err := a.PacketDecode(msg)
		if err != nil {
			a.session.Warnf("packet decoder error. [error = %s, msg = %s]", err, msg)
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

func (a *Agent) writeChan() {
	ticker := time.NewTicker(a.Heartbeat)
	defer func() {
		if clog.PrintLevel(zapcore.DebugLevel) {
			a.session.Debugf("close session. [sid = %s]", a.session.SID())
		}

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
				if clog.PrintLevel(zapcore.DebugLevel) {
					a.session.Debug("check heartbeat timeout.")
				}
				return
			}
		case bytes := <-a.chWrite:
			a.write(bytes)
		case pending := <-a.chSend:
			a.processMessage(pending)
		}
	}
}

func (a *Agent) write(bytes []byte) {
	_, err := a.conn.Write(bytes)
	if err != nil {
		clog.Warn(err)
	}
}

func (a *Agent) processMessage(data *pendingMessage) {
	payload, err := a.Marshal(data.payload)
	if err != nil {
		a.session.Warnf("payload marshal error. [data = %s]", data.String())
		return
	}

	// construct message and encode
	m := &cmsg.Message{
		Type:  data.typ,
		ID:    data.mid,
		Route: data.route,
		Data:  payload,
		Error: data.err,
	}

	// encode message
	em, err := cmsg.Encode(m)
	if err != nil {
		clog.Warn(err)
		return
	}

	// encode packet
	pkg, err := a.PacketEncode(cpkg.Data, em)
	if err != nil {
		clog.Warn(err)
		return
	}
	a.SendRaw(pkg)
}

func (a *Agent) processPacket(packet cfacade.IPacket) {
	result := a.session.OnDataListener()
	if result == false {
		if clog.PrintLevel(zapcore.WarnLevel) {
			a.session.Warnf("[ProcessPacket] on data listener return fail. [packet = %+v]", packet)
		}
		return
	}

	cmd, found := a.Commands[packet.Type()]
	if found == false {
		if clog.PrintLevel(zapcore.DebugLevel) {
			a.session.Debugf("[ProcessPacket] type not found. [packet = %+v]", packet)
		}
		return
	}

	cmd.Do(a.session, packet)

	// update last time
	a.SetLastAt()
}
