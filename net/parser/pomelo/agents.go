package pomelo

import (
	"sync"

	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

var (
	sidAgentMap = sync.Map{} // make(map[cfacade.SID]*Agent)      // sid -> Agent
	uidMap      = sync.Map{} // make(map[cfacade.UID]cfacade.SID) // uid -> sid
)

func BindSID(agent *Agent) {
	sidAgentMap.Store(agent.SID(), agent)
}

func Bind(sid cfacade.SID, uid cfacade.UID) (*Agent, error) {
	if sid == "" {
		return nil, cerr.Errorf("[sid = %s] less than 1.", sid)
	}

	if uid < 1 {
		return nil, cerr.Errorf("[uid = %d] less than 1.", uid)
	}

	// sid不存在，可能在执行该函数前已经断开连接
	agent, found := GetAgentWithSID(sid)
	if !found {
		return nil, cerr.Errorf("[sid = %s] does not exist.", sid)
	}

	// 先查找uid是否有旧的agent
	var oldAgent *Agent
	if oldSID, found := GetSID(uid); found && oldSID != sid {
		if agent, exists := GetAgentWithSID(oldSID); exists {
			oldAgent = agent
		}
	}

	// 再绑定uid
	agent.session.Uid = uid
	uidMap.Store(uid, sid)

	// 返回oldAgent(如果没有则为空，可自行处理，比如踢下线)
	return oldAgent, nil
}

func Unbind(sid cfacade.SID) {
	agent, found := GetAgentWithSIDAndDel(sid, true)
	if !found {
		return
	}

	// sid是自己，则删除uidmap
	if nowSID, ok := GetSID(agent.UID()); ok && nowSID == sid {
		uidMap.Delete(agent.UID())
	}

	clog.Debugf("Unbind agent. sid = %s", sid)
}

func GetAgentWithSIDAndDel(sid cfacade.SID, isDel bool) (*Agent, bool) {
	var (
		agentValue any
		found      bool
	)

	if isDel {
		agentValue, found = sidAgentMap.LoadAndDelete(sid)
	} else {
		agentValue, found = sidAgentMap.Load(sid)
	}

	if !found {
		return nil, false
	}

	agent, ok := agentValue.(*Agent)
	if !ok {
		return nil, false
	}

	return agent, found
}

func GetAgentWithSID(sid cfacade.SID) (*Agent, bool) {
	return GetAgentWithSIDAndDel(sid, false)
}

func GetAgentWithUID(uid cfacade.UID) (*Agent, bool) {
	if uid < 1 {
		return nil, false
	}

	sidValue, found := uidMap.Load(uid)
	if !found {
		return nil, false
	}

	sid := sidValue.(cfacade.UID)
	agentValue, found := sidAgentMap.Load(sid)
	if !found {
		return nil, false
	}

	agent, ok := agentValue.(*Agent)
	if !ok {
		return nil, false
	}

	return agent, found
}

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

func GetAgent(sid string, uid cfacade.UID) (*Agent, bool) {
	if sid != "" {
		return GetAgentWithSID(sid)
	}

	if uid > 0 {
		return GetAgentWithUID(uid)
	}

	return nil, false
}

func ForeachAgent(fn func(a *Agent)) {
	sidAgentMap.Range(func(key, value any) bool {
		if agent, ok := value.(*Agent); ok {
			fn(agent)
		}
		return true
	})
}

func Count() int {
	count := 0
	sidAgentMap.Range(func(key, value any) bool {
		count += 1
		return true
	})

	return count
}
