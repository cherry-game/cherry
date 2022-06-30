package cherrySession

import (
	cerr "github.com/cherry-game/cherry/error"
	cnuid "github.com/cherry-game/cherry/extend/nuid"
	cfacade "github.com/cherry-game/cherry/facade"
	"sync"
)

var (
	lock             = &sync.RWMutex{}
	sidMap           = make(map[cfacade.SID]*Session) // sid -> Session
	uidMap           = make(map[cfacade.UID]*Session) // uid -> Session
	onCreateListener = make([]SessionListener, 0)     // on create execute listener function
	onCloseListener  = make([]SessionListener, 0)     // on close execute listener function
	onDataListener   = make([]SessionListener, 0)     // on receive data execute listener function
)

type (
	SessionListener func(session *Session) (next bool)
)

func NextSID() cfacade.SID {
	return cnuid.Next()
}

func Create(sid cfacade.SID, frontendId cfacade.FrontendId, network cfacade.INetwork) *Session {
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

func Bind(sid cfacade.SID, uid cfacade.UID) error {
	if sid == "" {
		return cerr.Errorf("[sid = %d] less than 1.", sid)
	}

	if uid < 1 {
		return cerr.Errorf("[uid = %d] less than 1.", uid)
	}

	lock.Lock()
	defer lock.Unlock()

	session, found := sidMap[sid]
	if found == false {
		return cerr.Errorf("[sid = %s] does not exist.", sid)
	}

	if session.UID() > 0 && session.UID() == uid {
		return cerr.Errorf("[uid = %d] has already bound.", session.UID())
	}

	// set uid
	session.uid = uid
	// add uid map
	uidMap[uid] = session

	return nil
}

func Unbind(sid cfacade.SID) {
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

func GetBySID(sid cfacade.SID) (*Session, bool) {
	lock.RLock()
	defer lock.RUnlock()
	session, found := sidMap[sid]
	return session, found
}

func GetByUID(uid cfacade.UID) (*Session, bool) {
	lock.RLock()
	defer lock.RUnlock()
	session, found := uidMap[uid]
	return session, found
}

func Kick(uid cfacade.UID) error {
	if session, found := uidMap[uid]; found {
		session.Close()
		return nil
	}

	return cerr.Errorf("session does not exist, [uid = %d]", uid)
}

func KickBySID(sid cfacade.SID) error {
	if session, found := sidMap[sid]; found {
		session.Close()
		return nil
	}

	return cerr.Errorf("session does not exist, [sid = %d]", sid)
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
