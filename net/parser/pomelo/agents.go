package pomelo

import (
	"sync"

	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

var (
	lock        = &sync.RWMutex{}
	sidAgentMap = make(map[cfacade.SID]*Agent)      // sid -> Agent
	uidMap      = make(map[cfacade.UID]cfacade.SID) // uid -> sid
)

func BindSID(agent *Agent) {
	lock.Lock()
	defer lock.Unlock()

	sidAgentMap[agent.SID()] = agent
}

func Bind(sid cfacade.SID, uid cfacade.UID) (*Agent, error) {
	if sid == "" {
		return nil, cerr.Errorf("[sid = %s] less than 1.", sid)
	}

	if uid < 1 {
		return nil, cerr.Errorf("[uid = %d] less than 1.", uid)
	}

	lock.Lock()
	defer lock.Unlock()

	// sid不存在，可能在执行该函数前已经断开连接
	agent, found := sidAgentMap[sid]
	if !found {
		return nil, cerr.Errorf("[sid = %s] does not exist.", sid)
	}

	// 查找uid是否有旧的agent
	var oldAgent *Agent
	if oldsid, found := uidMap[uid]; found && oldsid != sid {
		if agent, exists := sidAgentMap[oldsid]; exists {
			oldAgent = agent
		}
	}

	// 绑定uid
	agent.session.Uid = uid
	uidMap[uid] = sid

	// 返回旧的agent(可自行处理，比如踢下线)
	return oldAgent, nil
}

func Unbind(sid cfacade.SID) {
	lock.Lock()
	defer lock.Unlock()

	agent, found := sidAgentMap[sid]
	if !found {
		return
	}

	delete(sidAgentMap, sid)
	delete(uidMap, agent.UID())

	sidCount := len(sidAgentMap)
	uidCount := len(uidMap)
	if sidCount == 0 || uidCount == 0 {
		clog.Infof("Unbind agent. sid = %s, sidCount = %d, uidCount = %d", sid, sidCount, uidCount)
	}
}

func GetAgent(sid cfacade.SID) (*Agent, bool) {
	lock.Lock()
	defer lock.Unlock()

	agent, found := sidAgentMap[sid]
	return agent, found
}

func GetAgentWithUID(uid cfacade.UID) (*Agent, bool) {
	if uid < 1 {
		return nil, false
	}

	lock.Lock()
	defer lock.Unlock()

	sid, found := uidMap[uid]
	if !found {
		return nil, false
	}

	agent, found := sidAgentMap[sid]
	return agent, found
}

func ForeachAgent(fn func(a *Agent)) {
	for _, agent := range sidAgentMap {
		fn(agent)
	}
}

func Count() int {
	lock.RLock()
	defer lock.RUnlock()

	return len(sidAgentMap)
}
