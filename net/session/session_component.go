package cherrySession

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	"sync"
)

type (
	//SessionComponent session component
	SessionComponent struct {
		sync.RWMutex
		facade.Component
		sidMap map[facade.SID]*Session // sid -> Session
		uidMap map[facade.UID]*Session // uid -> Session
	}
)

func NewComponent() *SessionComponent {
	s := &SessionComponent{
		sidMap: make(map[facade.SID]*Session),
		uidMap: make(map[facade.UID]*Session),
	}

	return s
}

func (s *SessionComponent) Name() string {
	return cherryConst.SessionComponent
}

func (s *SessionComponent) Create(sid facade.SID, frontendId facade.FrontendId) *Session {
	s.Lock()
	defer s.Unlock()

	session := NewSession(sid, frontendId)
	s.sidMap[session.sid] = session

	return session
}

func (s *SessionComponent) Bind(sid facade.SID, uid facade.UID) error {
	s.Lock()
	defer s.Unlock()

	session, found := s.sidMap[sid]
	if found == false {
		return cherryError.Errorf("sc does not exist, sid:%d", sid)
	}

	if session.UID() != "" {
		if session.UID() == uid {
			return nil
		}

		return cherryError.Errorf("sc has already bound with %s", session.UID())
	}

	err := session.Bind(uid)
	if err != nil {
		return err
	}

	s.uidMap[uid] = session

	return nil
}

func (s *SessionComponent) Remove(sid facade.SID) {
	s.Lock()
	defer s.Unlock()

	session, found := s.sidMap[sid]
	if found == false {
		return
	}

	delete(s.sidMap, sid)
	delete(s.uidMap, session.uid)
}

func (s *SessionComponent) Import(sid facade.SID, key string, value interface{}) error {
	if session, found := s.sidMap[sid]; found {
		session.Set(key, value)
		return nil
	}

	return cherryError.Errorf("session does not exist, sid:%d", sid)
}

func (s *SessionComponent) ImportAll(sid facade.SID, settings map[string]interface{}) error {
	if session, found := s.sidMap[sid]; found {
		for k, v := range settings {
			session.Set(k, v)
		}
		return nil
	}

	return cherryError.Errorf("session does not exist, sid:%d", sid)
}

func (s *SessionComponent) Kick(uid facade.UID) error {
	if session, found := s.uidMap[uid]; found {
		session.Closed()
		return nil
	}

	return cherryError.Errorf("session does not exist, uid:%s", uid)
}

func (s *SessionComponent) KickBySID(sid facade.SID) error {
	if session, found := s.sidMap[sid]; found {
		session.Closed()
		return nil
	}

	return cherryError.Errorf("session does not exist, sid:%d", sid)
}

func (s *SessionComponent) ForEachSIDSession(fn func(s *Session)) {
	for _, session := range s.sidMap {
		fn(session)
	}
}

func (s *SessionComponent) ForEachUIDSession(fn func(s *Session)) {
	for _, session := range s.uidMap {
		fn(session)
	}
}

func (s *SessionComponent) GetSessionCount() int {
	return len(s.sidMap)
}

func (s *SessionComponent) CloseAll() {
	for _, session := range s.uidMap {
		session.Closed()
	}
}
