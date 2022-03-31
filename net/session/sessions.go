package cherrySession

import (
	"github.com/cherry-game/cherry/error"
	cherryUUID "github.com/cherry-game/cherry/extend/uuid"
	facade "github.com/cherry-game/cherry/facade"
	"sync"
)

var (
	lock             = &sync.RWMutex{}
	sidMap           = make(map[facade.SID]*Session) // sid -> Session
	uidMap           = make(map[facade.UID]*Session) // uid -> Session
	onCreateListener = make([]SessionListener, 0)    // on create execute listener function
	onCloseListener  = make([]SessionListener, 0)    // on close execute listener function
	onDataListener   = make([]SessionListener, 0)    // on receive data execute listener function
)

type (
	SessionListener func(session *Session) (next bool)
)

func NextSID() facade.SID {
	return cherryUUID.New()
}

func Create(sid facade.SID, frontendId facade.FrontendId, network facade.INetwork) *Session {
	session := NewSession(sid, frontendId, network)

	lock.Lock()
	sidMap[session.sid] = session
	lock.Unlock()

	for _, listener := range onCreateListener {
		if listener(session) == false {
			break
		}
	}

	return session
}

func Bind(sid facade.SID, uid facade.UID) error {
	if sid == "" {
		return cherryError.Errorf("[sid = %d] less than 1.", sid)
	}

	if uid < 1 {
		return cherryError.Errorf("[uid = %d] less than 1.", uid)
	}

	lock.Lock()
	defer lock.Unlock()

	session, found := sidMap[sid]
	if found == false {
		return cherryError.Errorf("[sid = %s] does not exist.", sid)
	}

	if session.UID() > 0 && session.UID() == uid {
		return cherryError.Errorf("[uid = %d] has already bound.", session.UID())
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
	lock.RLock()
	defer lock.RUnlock()
	session, found := sidMap[sid]
	return session, found
}

func GetByUID(uid facade.UID) (*Session, bool) {
	lock.RLock()
	defer lock.RUnlock()
	session, found := uidMap[uid]
	return session, found
}

func Kick(uid facade.UID) error {
	if session, found := uidMap[uid]; found {
		session.Close()
		return nil
	}

	return cherryError.Errorf("session does not exist, [uid = %d]", uid)
}

func KickBySID(sid facade.SID) error {
	if session, found := sidMap[sid]; found {
		session.Close()
		return nil
	}

	return cherryError.Errorf("session does not exist, [sid = %d]", sid)
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

func CloseAll(cb func(session *Session)) {
	for _, session := range uidMap {
		cb(session)
		session.Close()
	}
}

func AddOnCreateListener(listener ...SessionListener) {
	onCreateListener = append(onCreateListener, listener...)
}

func AddOnCloseListener(listener ...SessionListener) {
	onCloseListener = append(onCloseListener, listener...)
}

func AddOnDataListener(listener ...SessionListener) {
	onDataListener = append(onDataListener, listener...)
}
