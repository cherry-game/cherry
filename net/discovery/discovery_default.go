package cherryDiscovery

import (
	cerr "github.com/cherry-game/cherry/error"
	cslice "github.com/cherry-game/cherry/extend/slice"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"sync"
)

// DiscoveryDefault 默认方式，通过读取profile文件的节点信息
//
// 该类型发现服务仅用于开发测试使用，直接读取profile.json->node配置
type DiscoveryDefault struct {
	sync.RWMutex
	memberList       []cfacade.IMember // key:nodeId,value:Member
	onAddListener    []cfacade.MemberListener
	onRemoveListener []cfacade.MemberListener
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

			nodeId := item.Get("node_id").ToString()
			if nodeId == "" {
				clog.Errorf("nodeId is empty in nodeType = %s", nodeType)
				break
			}

			if _, found := n.GetMember(nodeId); found {
				clog.Errorf("nodeType = %s, nodeId = %s, duplicate nodeId", nodeType, nodeId)
				break
			}

			member := &cproto.Member{
				NodeId:   nodeId,
				NodeType: nodeType,
				Address:  item.Get("rpc_address").ToString(),
				Settings: make(map[string]string),
			}

			settings := item.Get("__settings__")
			for _, key := range settings.Keys() {
				member.Settings[key] = settings.Get(key).ToString()
			}

			n.memberList = append(n.memberList, member)
		}
	}
}

func (n *DiscoveryDefault) Name() string {
	return "default"
}

func (n *DiscoveryDefault) List() []cfacade.IMember {
	return n.memberList
}

func (n *DiscoveryDefault) ListByType(nodeType string, filterNodeId ...string) []cfacade.IMember {
	var list []cfacade.IMember
	for _, member := range n.memberList {
		if member.GetNodeType() == nodeType {
			if _, ok := cslice.StringIn(member.GetNodeId(), filterNodeId); ok == false {
				list = append(list, member)
			}
		}
	}
	return list
}

func (n *DiscoveryDefault) GetType(nodeId string) (nodeType string, err error) {
	member, found := n.GetMember(nodeId)
	if found == false {
		return "", cerr.Errorf("nodeId = %s not found.", nodeId)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryDefault) GetMember(nodeId string) (cfacade.IMember, bool) {
	for _, member := range n.memberList {
		if member.GetNodeId() == nodeId {
			return member, true
		}
	}

	return nil, false
}

func (n *DiscoveryDefault) AddMember(member cfacade.IMember) {
	defer n.Unlock()
	n.Lock()

	if _, found := n.GetMember(member.GetNodeId()); found {
		clog.Warnf("duplicate nodeId. [nodeType = %s], [nodeId = %s], [address = %s]",
			member.GetNodeType(),
			member.GetNodeId(),
			member.GetAddress(),
		)
		return
	}

	n.memberList = append(n.memberList, &cproto.Member{
		NodeId:   member.GetNodeId(),
		NodeType: member.GetNodeType(),
		Address:  member.GetAddress(),
		Settings: member.GetSettings(),
	})

	for _, listener := range n.onAddListener {
		listener(member)
	}

	clog.Debugf("addMember new member. [member = %s]", member)
}

func (n *DiscoveryDefault) RemoveMember(nodeId string) {
	defer n.Unlock()
	n.Lock()

	if nodeId == "" {
		return
	}

	var member cfacade.IMember
	for i := 0; i < len(n.memberList); i++ {
		member = n.memberList[i]

		if member.GetNodeId() == nodeId {
			n.memberList = append(n.memberList[:i], n.memberList[i+1:]...)
			clog.Debugf("remove member. [member = %v]", member)

			for _, listener := range n.onRemoveListener {
				listener(member)
			}

			break
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
