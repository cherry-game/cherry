package cherrySession

import (
	"context"
	"fmt"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryContext "github.com/cherry-game/cherry/net/context"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
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
		state      int32             // current session state
		entity     facade.INetwork   // network
		sid        facade.SID        // session id
		uid        facade.UID        // user unique id
		frontendId facade.FrontendId // frontend node id
	}
)

func FakeSession(pbSession *cherryProto.Session, network facade.INetwork) *Session {
	session := &Session{
		settings: settings{
			data: make(map[string]string),
		},
		state:      Working,
		entity:     network,
		sid:        pbSession.Sid,
		uid:        pbSession.Uid,
		frontendId: pbSession.FrontendId,
	}
	session.ImportAll(pbSession.Data)

	return session
}

func NewSession(sid facade.SID, frontendId facade.FrontendId, network facade.INetwork) *Session {
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

func (s *Session) SID() facade.SID {
	return s.sid
}

func (s *Session) UID() facade.UID {
	return s.uid
}

func (s *Session) FrontendId() facade.FrontendId {
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
func (s *Session) RPC(route string, v interface{}, rsp *cherryProto.Response) {
	s.entity.RPC(route, v, rsp)
}

// Push message to client
func (s *Session) Push(route string, v interface{}) {
	s.entity.Push(route, v)
}

// ResponseMID responses message to client, mid is
// request message ID
func (s *Session) ResponseMID(mid uint, v interface{}, isError ...bool) {
	s.entity.Response(mid, v, isError...)

	if cherryProfile.Debug() {
		if len(isError) > 0 {
			s.Debugf("[ResponseMID] [mid = %d] [isError = %v] [data = %v]", mid, isError[0], v)
		} else {
			s.Debugf("[Response] [mid = %d] [data = %v]", mid, v)
		}
	}
}

func (s *Session) Response(ctx context.Context, v interface{}, isError ...bool) {
	mid := cherryContext.GetMessageId(ctx)
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

func (s *Session) OnCloseProcess() {
	// when session closed,the func is executed
	for _, listener := range onCloseListener {
		if listener(s) == false {
			break
		}
	}
	Unbind(s.sid)
}

func (s *Session) RemoteAddress() string {
	if s.entity == nil {
		return ""
	}
	return s.entity.RemoteAddr()
}

func (s *Session) String() string {
	return fmt.Sprintf("sid = %s, uid = %d, address = %s",
		s.sid,
		s.uid,
		s.RemoteAddress(),
	)
}

func (s *Session) logPrefix() string {
	return fmt.Sprintf("[uid = %d] ", s.uid)
}

func (s *Session) Debug(args ...interface{}) {
	cherryLogger.DefaultLogger.Debug(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Debugf(template string, args ...interface{}) {
	cherryLogger.DefaultLogger.Debug(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) Info(args ...interface{}) {
	cherryLogger.DefaultLogger.Info(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Infof(template string, args ...interface{}) {
	cherryLogger.DefaultLogger.Info(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) Warn(args ...interface{}) {
	cherryLogger.DefaultLogger.Warn(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Warnf(template string, args ...interface{}) {
	cherryLogger.DefaultLogger.Warn(s.logPrefix(), fmt.Sprintf(template, args...))
}

func (s *Session) Error(args ...interface{}) {
	cherryLogger.DefaultLogger.Error(s.logPrefix(), fmt.Sprint(args...))
}

func (s *Session) Errorf(template string, args ...interface{}) {
	cherryLogger.DefaultLogger.Error(s.logPrefix(), fmt.Sprintf(template, args...))
}
