package cherrySession

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"net"
	"sync/atomic"
)

var (
	INIT   = 0
	CLOSED = 1

	atomicSessionId int64
)

func NextSessionId() int64 {
	return atomic.AddInt64(&atomicSessionId, 1)
}

//SessionComponent session sessionComponent
type SessionComponent struct {
	cherryInterfaces.BaseComponent
	sidMap map[cherryInterfaces.SID]*Session   // sid -> Session
	uidMap map[cherryInterfaces.UID][]*Session // uid -> Session
}

func NewService() *SessionComponent {
	return &SessionComponent{
		sidMap: make(map[cherryInterfaces.SID]*Session),
		uidMap: make(map[cherryInterfaces.UID][]*Session),
	}
}

func (s *SessionComponent) Name() string {
	return cherryConst.SessionComponent
}

func (s *SessionComponent) Create(conn net.Conn, net cherryInterfaces.INetworkEntity) *Session {
	newSession := NewSession(NextSessionId(), conn, net, s)
	s.sidMap[newSession.SID()] = newSession
	return newSession
}

func (s *SessionComponent) Bind(sid cherryInterfaces.SID, uid cherryInterfaces.UID) error {
	session := s.sidMap[sid]

	if session == nil {
		return cherryUtils.Errorf("sessionComponent does not exist, sid:%d", sid)
	}

	if session.UID() < 1 && session.UID() == uid {
		return cherryUtils.Errorf("sessionComponent has already bound with %d", session.UID())
	}

	sessions := s.uidMap[uid]

	for _, s := range sessions {
		if s.SID() == session.SID() {
			return nil
		}
	}

	return nil
}

func (s *SessionComponent) Remove(sid cherryInterfaces.SID) {
	delete(s.sidMap, sid)
}
