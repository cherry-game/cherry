package cherrySession

import (
	"github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	"sync"
	"sync/atomic"
)

var (
	nextSessionId    int64
	lock             = &sync.RWMutex{}
	sidMap           = make(map[facade.SID]*Session) // sid -> Session
	uidMap           = make(map[facade.UID]*Session) // uid -> Session
	onCreateListener = make([]SessionListener, 0)    // on create execute listener function
	onCloseListener  = make([]SessionListener, 0)    // on close execute listener function
)

type (
	SessionListener func(session *Session) (next bool)
)

func NextSID() facade.SID {
	return atomic.AddInt64(&nextSessionId, 1)
}

func Create(nodeId string, entity facade.INetwork) *Session {
	lock.Lock()
	defer lock.Unlock()

	session := NewSession(NextSID(), nodeId, entity)

	sidMap[session.sid] = session

	return session
}

func Bind(sid facade.SID, uid facade.UID) error {
	if sid < 1 {
		return cherryError.Errorf("sid[%d] less than 1.", sid)
	}

	if uid < 1 {
		return cherryError.Errorf("uid[%d] less than 1.", uid)
	}

	lock.Lock()
	defer lock.Unlock()

	session, found := sidMap[sid]
	if found == false {
		return cherryError.Errorf("sid[%d] does not exist.", sid)
	}

	if session.UID() > 0 && session.UID() == uid {
		return cherryError.Errorf("uid[%d] has already bound.", session.UID())
	}

	// set uid
	session.uid = uid
	// add uid map
	uidMap[uid] = session

	return nil
}

func Unbind(sid facade.SID) {
	lock.Lock()
	defer lock.Unlock()

	session, found := sidMap[sid]
	if found == false {
		return
	}

	delete(sidMap, sid)
	delete(uidMap, session.uid)

	session.uid = 0
}

func GetBySID(sid facade.SID) (*Session, bool) {
	session, found := sidMap[sid]
	return session, found
}

func GetByUID(uid facade.UID) (*Session, bool) {
	session, found := uidMap[uid]
	return session, found
}

func Import(sid facade.SID, key string, value interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	if session, found := sidMap[sid]; found {
		session.Set(key, value)
		return nil
	}

	return cherryError.Errorf("session does not exist, sid[%d]", sid)
}

func ImportAll(sid facade.SID, settings map[string]interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	if session, found := sidMap[sid]; found {
		for k, v := range settings {
			session.Set(k, v)
		}
		return nil
	}

	return cherryError.Errorf("session does not exist, sid[%d]", sid)
}

func Kick(uid facade.UID) error {
	if session, found := uidMap[uid]; found {
		session.Close()
		return nil
	}

	return cherryError.Errorf("session does not exist, uid[%d]", uid)
}

func KickBySID(sid facade.SID) error {
	if session, found := sidMap[sid]; found {
		session.Close()
		return nil
	}

	return cherryError.Errorf("session does not exist, sid[%d]", sid)
}

func ForEachSIDSession(fn func(s *Session)) {
	for _, session := range sidMap {
		fn(session)
	}
}

func ForEachUIDSession(fn func(s *Session)) {
	for _, session := range uidMap {
		fn(session)
	}
}

func GetSessionCount() int {
	return len(sidMap)
}

func CloseAll() {
	for _, session := range uidMap {
		session.Close()
	}
}

func AddOnCreateListener(listener ...SessionListener) {
	onCreateListener = append(onCreateListener, listener...)
}

func AddOnCloseListener(listener ...SessionListener) {
	onCloseListener = append(onCloseListener, listener...)
}
