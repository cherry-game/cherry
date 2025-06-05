package cherryDiscovery

import (
	"math/rand"
	"sync"

	cerr "github.com/cherry-game/cherry/error"
	cslice "github.com/cherry-game/cherry/extend/slice"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
)

// DiscoveryDefault 默认方式，通过读取profile文件的节点信息
//
// 该类型发现服务仅用于开发测试使用，直接读取profile.json->node配置
type DiscoveryDefault struct {
	memberMap        sync.Map // key:nodeID,value:cfacade.IMember
	onAddListener    []cfacade.MemberListener
	onRemoveListener []cfacade.MemberListener
}

func (n *DiscoveryDefault) PreInit() {
	n.memberMap = sync.Map{}
}

func (n *DiscoveryDefault) Load(_ cfacade.IApplication) {
	// load node info from profile file
	nodeConfig := cprofile.GetConfig("node")
	if nodeConfig.LastError() != nil {
		clog.Error("`node` property not found in profile file.")
		return
	}

	for _, nodeType := range nodeConfig.Keys() {
		typeJson := nodeConfig.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)

			nodeID := item.Get("node_id").ToString()
			if nodeID == "" {
				clog.Errorf("nodeID is empty in nodeType = %s", nodeType)
				break
			}

			if _, found := n.GetMember(nodeID); found {
				clog.Errorf("nodeType = %s, nodeID = %s, duplicate nodeID", nodeType, nodeID)
				break
			}

			member := &cproto.Member{
				NodeID:   nodeID,
				NodeType: nodeType,
				Address:  item.Get("rpc_address").ToString(),
				Settings: make(map[string]string),
			}

			settings := item.Get("__settings__")
			for _, key := range settings.Keys() {
				member.Settings[key] = settings.Get(key).ToString()
			}

			n.memberMap.Store(member.NodeID, member)
		}
	}
}

func (n *DiscoveryDefault) Name() string {
	return "default"
}

func (n *DiscoveryDefault) Map() map[string]cfacade.IMember {
	memberMap := map[string]cfacade.IMember{}

	n.memberMap.Range(func(key, value any) bool {
		if member, ok := value.(cfacade.IMember); ok {
			memberMap[member.GetNodeID()] = member
		}
		return true
	})

	return memberMap
}

func (n *DiscoveryDefault) ListByType(nodeType string, filterNodeID ...string) []cfacade.IMember {
	var memberList []cfacade.IMember

	n.memberMap.Range(func(key, value any) bool {
		member := value.(cfacade.IMember)
		if member.GetNodeType() == nodeType {
			if _, ok := cslice.StringIn(member.GetNodeID(), filterNodeID); !ok {
				memberList = append(memberList, member)
			}
		}

		return true
	})

	return memberList
}

func (n *DiscoveryDefault) Random(nodeType string) (cfacade.IMember, bool) {
	memberList := n.ListByType(nodeType)
	memberLen := len(memberList)

	if memberLen < 1 {
		return nil, false
	}

	if memberLen == 1 {
		return memberList[0], true
	}

	return memberList[rand.Intn(len(memberList))], true
}

func (n *DiscoveryDefault) GetType(nodeID string) (nodeType string, err error) {
	member, found := n.GetMember(nodeID)
	if !found {
		return "", cerr.Errorf("nodeID = %s not found.", nodeID)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryDefault) GetMember(nodeID string) (cfacade.IMember, bool) {
	if nodeID == "" {
		return nil, false
	}

	value, found := n.memberMap.Load(nodeID)
	if !found {
		return nil, false
	}

	return value.(cfacade.IMember), found
}

func (n *DiscoveryDefault) AddMember(member cfacade.IMember) {
	_, loaded := n.memberMap.LoadOrStore(member.GetNodeID(), member)
	if loaded {
		clog.Warnf("duplicate nodeID. [nodeType = %s], [nodeID = %s], [address = %s]",
			member.GetNodeType(),
			member.GetNodeID(),
			member.GetAddress(),
		)
		return
	}

	for _, listener := range n.onAddListener {
		listener(member)
	}

	clog.Debugf("addMember new member. [member = %s]", member)
}

func (n *DiscoveryDefault) RemoveMember(nodeID string) {
	value, loaded := n.memberMap.LoadAndDelete(nodeID)
	if loaded {
		member := value.(cfacade.IMember)
		clog.Debugf("remove member. [member = %s]", member)

		for _, listener := range n.onRemoveListener {
			listener(member)
		}
	}
}

func (n *DiscoveryDefault) OnAddMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onAddListener = append(n.onAddListener, listener)
}

func (n *DiscoveryDefault) OnRemoveMember(listener cfacade.MemberListener) {
	if listener == nil {
		return
	}
	n.onRemoveListener = append(n.onRemoveListener, listener)
}

func (n *DiscoveryDefault) Stop() {

}
