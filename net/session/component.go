package cherrySession

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	"sync"
)

type (
	//Component session component
	Component struct {
		sync.RWMutex
		facade.Component
		sidMap   map[facade.SID]*Session // sid -> Session
		uidMap   map[facade.UID]*Session // uid -> Session
		onCreate []SessionListener       // on create execute listener function
		onClose  []SessionListener       // on close execute listener function
	}

	SessionListener func(session *Session) (next bool)
)

func NewComponent() *Component {
	s := &Component{
		sidMap:   make(map[facade.SID]*Session),
		uidMap:   make(map[facade.UID]*Session),
		onCreate: make([]SessionListener, 0),
		onClose:  make([]SessionListener, 0),
	}

	return s
}

func (s *Component) Name() string {
	return cherryConst.SessionComponent
}

func (s *Component) Create(entity facade.INetwork) *Session {
	s.Lock()
	defer s.Unlock()

	session := NewSession(NextSID(), s.App().NodeId(), entity, s)

	s.sidMap[session.sid] = session

	return session
}

func (s *Component) Bind(sid facade.SID, uid facade.UID) error {
	s.Lock()
	defer s.Unlock()

	session, found := s.sidMap[sid]
	if found == false {
		return cherryError.Errorf("sc does not exist, sid:%d", sid)
	}

	if session.UID() > 0 {
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

func (s *Component) Remove(sid facade.SID) {
	s.Lock()
	defer s.Unlock()

	session, found := s.sidMap[sid]
	if found == false {
		return
	}

	delete(s.sidMap, sid)
	delete(s.uidMap, session.uid)
}

func (s *Component) Import(sid facade.SID, key string, value interface{}) error {
	if session, found := s.sidMap[sid]; found {
		session.Set(key, value)
		return nil
	}

	return cherryError.Errorf("session does not exist, sid:%d", sid)
}

func (s *Component) ImportAll(sid facade.SID, settings map[string]interface{}) error {
	if session, found := s.sidMap[sid]; found {
		for k, v := range settings {
			session.Set(k, v)
		}
		return nil
	}

	return cherryError.Errorf("session does not exist, sid:%d", sid)
}

func (s *Component) Kick(uid facade.UID) error {
	if session, found := s.uidMap[uid]; found {
		session.Closed()
		return nil
	}

	return cherryError.Errorf("session does not exist, uid:%s", uid)
}

func (s *Component) KickBySID(sid facade.SID) error {
	if session, found := s.sidMap[sid]; found {
		session.Closed()
		return nil
	}

	return cherryError.Errorf("session does not exist, sid:%d", sid)
}

func (s *Component) ForEachSIDSession(fn func(s *Session)) {
	for _, session := range s.sidMap {
		fn(session)
	}
}

func (s *Component) ForEachUIDSession(fn func(s *Session)) {
	for _, session := range s.uidMap {
		fn(session)
	}
}

func (s *Component) GetSessionCount() int {
	return len(s.sidMap)
}

func (s *Component) CloseAll() {
	for _, session := range s.uidMap {
		session.Closed()
	}
}

func (s *Component) AddOnCreate(listener ...SessionListener) {
	s.onCreate = append(s.onCreate, listener...)
}

func (s *Component) AddOnClose(listener ...SessionListener) {
	s.onClose = append(s.onClose, listener...)
}
