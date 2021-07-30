package cherrySession

import (
	"fmt"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
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

func NewSession(sid facade.SID, frontendId facade.FrontendId, entity facade.INetwork) *Session {
	session := &Session{
		settings: settings{
			data: make(map[string]interface{}),
		},
		entity:     entity,
		sid:        sid,
		uid:        0,
		frontendId: frontendId,
	}

	for _, listener := range onCreateListener {
		if listener(session) == false {
			break
		}
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

func (s *Session) SendRaw(bytes []byte) error {
	if s.entity == nil {
		s.Debug("entity is nil. bytes = %v", bytes)
		return nil
	}

	return s.entity.SendRaw(bytes)
}

// RPC sends message to remote server
func (s *Session) RPC(route string, val interface{}) error {
	if s.entity == nil {
		s.Debug("entity is nil. route = %s, val = %v", route, val)
		return nil
	}

	return s.entity.RPC(route, val)
}

// Push message to client
func (s *Session) Push(route string, val interface{}) error {
	if s.entity == nil {
		s.Debug("entity is nil. route = %s, val = %v", route, val)
		return nil
	}

	return s.entity.Push(route, val)
}

// Response responses message to client, mid is
// request message ID
func (s *Session) Response(mid uint, val interface{}, isError ...bool) error {
	if s.entity == nil {
		s.Debug("entity is nil. mid = %d, val = %v, isError = %v", mid, val, isError)
		return nil
	}

	return s.entity.Response(mid, val, isError...)
}

func (s *Session) Kick(reason interface{}, close bool) {
	if s.entity == nil {
		s.Debug("entity is nil. reason = %v, close = %v", reason, close)
		return
	}

	err := s.entity.Kick(reason)
	if err != nil {
		s.Warn(err)
		return
	}

	if close {
		s.Close()
	}
}

func (s *Session) Close() {
	if s.entity == nil {
		s.Debug("entity is nil")
		return
	}

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
		s.Debug("entity is nil")
		return ""
	}

	return s.entity.RemoteAddr().String()
}

func (s *Session) String() string {
	return fmt.Sprintf("sid = %d, uid = %d, address = %s",
		s.sid,
		s.uid,
		s.RemoteAddress(),
	)
}

func (s *Session) logPrefix() string {
	return fmt.Sprintf("[sid=%d, uid=%d] ", s.sid, s.uid)
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
