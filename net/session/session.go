package cherrySession

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	"sync/atomic"
)

var nextSessionId int64

func NextSID() facade.SID {
	return atomic.AddInt64(&nextSessionId, 1)
}

type (
	Session struct {
		Settings
		entity     facade.INetwork   // network
		sid        facade.SID        // session id
		uid        facade.UID        // user unique id
		frontendId facade.FrontendId // frontend node id
	}
)

func NewSession(sid facade.SID, frontendId facade.FrontendId) *Session {
	session := &Session{
		Settings: Settings{
			data: make(map[string]interface{}),
		},
		sid:        sid,
		uid:        "",
		frontendId: frontendId,
	}

	return session
}

func (s *Session) SetNetwork(entity facade.INetwork) {
	s.entity = entity
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

func (s *Session) Bind(uid facade.UID) error {
	if uid == "" {
		return cherryError.SessionIllegalUID
	}

	s.Lock()
	defer s.Unlock()

	s.uid = uid

	return nil
}

func (s *Session) Unbind() {
	s.Lock()
	defer s.Unlock()

	s.uid = ""
}

func (s *Session) SendRaw(bytes []byte) error {
	return s.entity.SendRaw(bytes)
}

// RPC sends message to remote server
func (s *Session) RPC(route string, v interface{}) error {
	return s.entity.RPC(route, v)
}

// Push message to client
func (s *Session) Push(route string, v interface{}) error {
	return s.entity.Push(route, v)
}

// ResponseMID responses message to client, mid is
// request message ID
func (s *Session) ResponseMID(mid uint64, v interface{}) error {
	return s.entity.ResponseMid(mid, v)
}

func (s *Session) Closed() {
	s.entity.Close()
}

func (s *Session) RemoteAddress() string {
	return s.entity.RemoteAddr().String()
}

func (s *Session) String() string {
	return fmt.Sprintf("sid = %d, uid = %s, address = %s",
		s.sid,
		s.uid,
		s.RemoteAddress(),
	)
}
