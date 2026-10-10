package pomelo

import (
	"sync"

	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

var (
	sidAgentMap = sync.Map{} // sid → *Agent
	uidMap      = sync.Map{} // uid → sid
)

// BindSID registers an agent in the sid lookup map.
func BindSID(agent *Agent) {
	sidAgentMap.Store(agent.SID(), agent)
}

// Bind associates a uid with an existing sid.
// Returns any previously bound agent for the same uid (duplicate login).
func Bind(sid cfacade.SID, uid cfacade.UID) (*Agent, error) {
	if sid == "" {
		return nil, cerr.Errorf("[sid = %s] less than 1.", sid)
	}
	if uid < 1 {
		return nil, cerr.Errorf("[uid = %d] less than 1.", uid)
	}

	agent, found := GetAgentWithSID(sid)
	if !found {
		return nil, cerr.Errorf("[sid = %s] does not exist.", sid)
	}

	var oldAgent *Agent
	if oldSID, found := GetSID(uid); found && oldSID != sid {
		oldAgent, _ = GetAgentWithSID(oldSID)
	}

	agent.SetUID(uid)
	uidMap.Store(uid, sid)

	return oldAgent, nil
}

// Unbind removes the sid→agent mapping and cleans up the uid→sid entry
// if it still points to this sid.
func Unbind(sid cfacade.SID) {
	agent, found := popAgent(sid)
	if !found {
		return
	}

	if nowSID, ok := GetSID(agent.UID()); ok && nowSID == sid {
		uidMap.Delete(agent.UID())
	}

	clog.Debugf("Unbind agent. sid = %s", sid)
}

// popAgent looks up and removes an agent by sid.
func popAgent(sid cfacade.SID) (*Agent, bool) {
	agentValue, found := sidAgentMap.LoadAndDelete(sid)
	if !found {
		return nil, false
	}
	return agentValue.(*Agent), true
}

// GetAgentWithSID looks up an agent by session id.
func GetAgentWithSID(sid cfacade.SID) (*Agent, bool) {
	agentValue, found := sidAgentMap.Load(sid)
	if !found {
		return nil, false
	}
	return agentValue.(*Agent), true
}

// GetAgentWithUID looks up an agent by user id.
func GetAgentWithUID(uid cfacade.UID) (*Agent, bool) {
	if uid < 1 {
		return nil, false
	}
	sid, found := GetSID(uid)
	if !found {
		return nil, false
	}
	return GetAgentWithSID(sid)
}

// GetAgent resolves an agent by sid first, falling back to uid.
func GetAgent(sid string, uid cfacade.UID) (*Agent, bool) {
	if sid != "" {
		return GetAgentWithSID(sid)
	}
	if uid > 0 {
		return GetAgentWithUID(uid)
	}
	return nil, false
}

// GetSID returns the session id bound to the given user id.
func GetSID(uid int64) (cfacade.SID, bool) {
	sidValue, found := uidMap.Load(uid)
	if !found {
		return "", false
	}
	sid, ok := sidValue.(cfacade.SID)
	if !ok {
		return "", false
	}
	return sid, true
}

// ForeachAgent iterates over all connected agents.
func ForeachAgent(fn func(a *Agent)) {
	sidAgentMap.Range(func(key, value any) bool {
		if agent, ok := value.(*Agent); ok {
			fn(agent)
		}
		return true
	})
}

// Count returns the total number of connected agents.
func Count() int {
	count := 0
	sidAgentMap.Range(func(key, value any) bool {
		count += 1
		return true
	})
	return count
}
