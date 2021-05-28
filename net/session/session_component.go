package cherrySession

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/facade"
	"net"
	"sync/atomic"
)

var (
	atomicSessionId int64
)

func NewSessionId() int64 {
	return atomic.AddInt64(&atomicSessionId, 1)
}

//SessionComponent session sessionComponent
type SessionComponent struct {
	cherryFacade.Component
	sidMap map[cherryFacade.SID]*Session   // sid -> Session
	uidMap map[cherryFacade.UID][]*Session // uid -> Session
}

func NewService() *SessionComponent {
	return &SessionComponent{
		sidMap: make(map[cherryFacade.SID]*Session),
		uidMap: make(map[cherryFacade.UID][]*Session),
	}
}

func (s *SessionComponent) Name() string {
	return cherryConst.SessionComponent
}

func (s *SessionComponent) Create(conn net.Conn, net cherryFacade.INetworkEntity) *Session {
	newSession := NewSession(NewSessionId(), conn, net, s)
	s.sidMap[newSession.SID()] = newSession
	return newSession
}

func (s *SessionComponent) Bind(sid cherryFacade.SID, uid cherryFacade.UID) error {
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

func (s *SessionComponent) Remove(sid cherryFacade.SID) {
	delete(s.sidMap, sid)
}
