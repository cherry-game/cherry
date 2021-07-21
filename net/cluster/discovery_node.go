package cherryCluster

import (
	cherryError "github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	jsoniter "github.com/json-iterator/go"
)

// DiscoveryNode 读取配置类型节点发现
//
// 该类型发现服务仅用于开发测试使用，直接读取profile.json->node配置
type DiscoveryNode struct {
	memberList []*cherryProto.Member // key:nodeId,value:Member
}

func (n *DiscoveryNode) Name() string {
	return "node"
}

func (n *DiscoveryNode) Init(_ facade.IApplication, _ jsoniter.Any) {
	nodes := cherryProfile.Config().Get(n.Name())
	if nodes.LastError() != nil {
		cherryLogger.Error("`nodes` property not found in profile file.")
		return
	}

	for _, nodeType := range nodes.Keys() {
		typeJson := nodes.Get(nodeType)
		for i := 0; i < typeJson.Size(); i++ {
			item := typeJson.Get(i)

			nodeId := item.Get("node_id").ToString()
			if nodeId == "" {
				cherryLogger.Errorf("nodeId is empty in nodeType = %s", nodeType)
				break
			}

			if _, found := n.Get(nodeId); found {
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

func (n *DiscoveryNode) List() []facade.IMember {
	var list []facade.IMember
	for _, member := range n.memberList {
		list = append(list, member)
	}
	return list
}

func (n *DiscoveryNode) GetType(nodeId string) (nodeType string, err error) {
	member, found := n.Get(nodeId)
	if found == false {
		return "", cherryError.Errorf("nodeId = %s not found.", nodeId)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryNode) Get(nodeId string) (facade.IMember, bool) {
	for _, member := range n.memberList {
		if member.GetNodeId() == nodeId {
			return member, true
		}
	}

	return nil, false
}

func (n *DiscoveryNode) OnAddMember(_ facade.MemberListener) {
	cherryLogger.Debug("No implementation required")
}

func (n *DiscoveryNode) OnRemoveMember(_ facade.MemberListener) {
	cherryLogger.Debug("No implementation required")
}
