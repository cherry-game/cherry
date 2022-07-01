package cherrySession

import (
	"context"
	"fmt"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	ccontext "github.com/cherry-game/cherry/net/context"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap/zapcore"
	"sync/atomic"
)

const (
	Init = iota
	WaitAck
	Working
	Closed
)

type (
	Session struct {
		settings
		state      int32              // current session state
		entity     cfacade.INetwork   // network
		sid        cfacade.SID        // session id
		uid        cfacade.UID        // user unique id
		frontendId cfacade.FrontendId // frontend node id
	}
)

func FakeSession(request *cproto.Request, network cfacade.INetwork) *Session {
	session := &Session{
		settings: settings{
			data: make(map[string]string),
		},
		state:      Working,
		entity:     network,
		sid:        request.Sid,
		uid:        request.Uid,
		frontendId: request.FrontendId,
	}
	session.ImportAll(request.Setting)

	return session
}

func NewSession(sid cfacade.SID, frontendId cfacade.FrontendId, network cfacade.INetwork) *Session {
	session := &Session{
		settings: settings{
			data: make(map[string]string),
		},
		state:      Init,
		entity:     network,
		sid:        sid,
		uid:        0,
		frontendId: frontendId,
	}

	return session
}

func (s *Session) State() int32 {
	return atomic.LoadInt32(&s.state)
}

func (s *Session) SetState(state int32) {
	atomic.StoreInt32(&s.state, state)
}

func (s *Session) SID() cfacade.SID {
	return s.sid
}

func (s *Session) UID() cfacade.UID {
	return s.uid
}

func (s *Session) FrontendId() cfacade.FrontendId {
	return s.frontendId
}

func (s *Session) IsBind() bool {
	return s.uid > 0
}

func (s *Session) SendRaw(bytes []byte) {
	if s.entity == nil {
		s.Debug("entity is nil")
		return
	}

	s.entity.SendRaw(bytes)
}

// RPC sends message to remote server
func (s *Session) RPC(nodeId string, route string, req proto.Message, rsp proto.Message) int32 {
	return s.entity.RPC(nodeId, route, req, rsp)
}

// Push message to client
func (s *Session) Push(route string, v interface{}) {
	s.entity.Push(route, v)
}

// ResponseMID responses message to client, mid is
// request message ID
func (s *Session) ResponseMID(mid uint, v interface{}, isError ...bool) {
	s.entity.Response(mid, v, isError...)
}

func (s *Session) Response(ctx context.Context, v interface{}, isError ...bool) {
	mid := ccontext.GetMessageId(ctx)
	s.ResponseMID(mid, v, isError...)
}

func (s *Session) Kick(reason interface{}, close bool) {
	s.entity.Kick(reason)

	if close {
		s.Close()
	}
}

func (s *Session) Close() {
	s.entity.Close()
}

func (s *Session) OnCloseListener() {
	// when session closed,the func is executed
	for _, listener := range onCloseListener {
		if listener(s) == false {
			break
		}
	}
	Unbind(s.sid)
}

func (s *Session) OnDataListener() bool {
	for _, listener := range onDataListener {
		if listener(s) == false {
			return false
		}
	}

	return true
}

func (s *Session) RemoteAddress() string {
	if s.entity == nil {
		return ""
	}
	return s.entity.RemoteAddr()
}

func (s *Session) String() string {
	return fmt.Sprintf("[sid = %s, uid = %d, address = %s]",
		s.sid,
		s.uid,
		s.RemoteAddress(),
	)
}

func (s *Session) logPrefix() string {
	return fmt.Sprintf("[uid = %d] ", s.uid)
}

func (s *Session) Debug(args ...interface{}) {
	clog.DefaultLogger.Debug(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Debugf(template string, args ...interface{}) {
	clog.DefaultLogger.Debug(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) Info(args ...interface{}) {
	clog.DefaultLogger.Info(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Infof(template string, args ...interface{}) {
	clog.DefaultLogger.Info(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) Warn(args ...interface{}) {
	clog.DefaultLogger.Warn(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Warnf(template string, args ...interface{}) {
	clog.DefaultLogger.Warn(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) Error(args ...interface{}) {
	clog.DefaultLogger.Error(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Errorf(template string, args ...interface{}) {
	clog.DefaultLogger.Error(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) LogEnable(level zapcore.Level) bool {
	return clog.DefaultLogger.Enable(level)
}
