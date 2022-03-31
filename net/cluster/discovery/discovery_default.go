package cherryDiscovery

import (
	cherryError "github.com/cherry-game/cherry/error"
	cherrySlice "github.com/cherry-game/cherry/extend/slice"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"sync"
)

// DiscoveryDefault 默认方式，读取profile文件
//
// 该类型发现服务仅用于开发测试使用，直接读取profile.json->node配置
type DiscoveryDefault struct {
	sync.RWMutex
	memberList       []*cherryProto.Member // key:nodeId,value:Member
	onAddListener    []facade.MemberListener
	onRemoveListener []facade.MemberListener
}

func (n *DiscoveryDefault) Name() string {
	return "default"
}

func (n *DiscoveryDefault) Init(_ facade.IApplication) {

	// load node info from profile file
	nodeProfile := cherryProfile.Get("node")
	if nodeProfile.LastError() != nil {
		cherryLogger.Error("`node` property not found in profile file.")
		return
	}

	for _, nodeType := range nodeProfile.Keys() {
		typeJson := nodeProfile.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)

			nodeId := item.Get("node_id").ToString()
			if nodeId == "" {
				cherryLogger.Errorf("nodeId is empty in nodeType = %s", nodeType)
				break
			}

			if _, found := n.GetMember(nodeId); found {
				cherryLogger.Errorf("nodeType = %s, nodeId = %s, duplicate nodeId", nodeType, nodeId)
				break
			}

			member := &cherryProto.Member{
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

func (n *DiscoveryDefault) List() []facade.IMember {
	var list []facade.IMember
	for _, member := range n.memberList {
		list = append(list, member)
	}
	return list
}

func (n *DiscoveryDefault) ListByType(nodeType string, filterNodeId ...string) []facade.IMember {
	var list []facade.IMember
	for _, member := range n.memberList {
		if member.GetNodeType() == nodeType {
			if _, ok := cherrySlice.StringIn(member.GetNodeId(), filterNodeId); ok == false {
				list = append(list, member)
			}
		}
	}
	return list
}

func (n *DiscoveryDefault) GetType(nodeId string) (nodeType string, err error) {
	member, found := n.GetMember(nodeId)
	if found == false {
		return "", cherryError.Errorf("nodeId = %s not found.", nodeId)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryDefault) GetMember(nodeId string) (facade.IMember, bool) {
	for _, member := range n.memberList {
		if member.GetNodeId() == nodeId {
			return member, true
		}
	}

	return nil, false
}

func (n *DiscoveryDefault) AddMember(member facade.IMember) {
	defer n.Unlock()
	n.Lock()

	if _, found := n.GetMember(member.GetNodeId()); found {
		cherryLogger.Warnf("duplicate nodeId. [nodeType = %s], [nodeId = %s], [fromAddress = %s]",
			member.GetNodeType(),
			member.GetNodeId(),
			member.GetAddress(),
		)
		return
	}

	n.memberList = append(n.memberList, &cherryProto.Member{
		NodeId:   member.GetNodeId(),
		NodeType: member.GetNodeType(),
		Address:  member.GetAddress(),
		Settings: member.GetSettings(),
	})

	for _, listener := range n.onAddListener {
		listener(member)
	}

	cherryLogger.Debugf("addMember new member. [member = %s]", member)
}

func (n *DiscoveryDefault) RemoveMember(nodeId string) {
	defer n.Unlock()
	n.Lock()

	if nodeId == "" {
		return
	}

	var member facade.IMember
	for i := 0; i < len(n.memberList); i++ {
		member = n.memberList[i]

		if member.GetNodeId() == nodeId {
			n.memberList = append(n.memberList[:i], n.memberList[i+1:]...)
			cherryLogger.Debugf("remove member. [member = %v]", member)

			for _, listener := range n.onRemoveListener {
				listener(member)
			}

			break
		}
	}
}

func (n *DiscoveryDefault) OnAddMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	n.onAddListener = append(n.onAddListener, listener)
}

func (n *DiscoveryDefault) OnRemoveMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	n.onRemoveListener = append(n.onRemoveListener, listener)
}

func (n *DiscoveryDefault) OnStop() {

}
