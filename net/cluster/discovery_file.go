package cherryCluster

import (
	cherryError "github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"
	"sync"
)

// DiscoveryFile 读取文件配置模式
//
// 该类型发现服务仅用于开发测试使用，直接读取profile.json->node配置
type DiscoveryFile struct {
	sync.RWMutex
	memberList       []*cherryProto.Member // key:nodeId,value:Member
	onAddListener    []facade.MemberListener
	onRemoveListener []facade.MemberListener
}

func (n *DiscoveryFile) Name() string {
	return "file"
}

func (n *DiscoveryFile) Init(_ facade.IApplication, _ *grpc.Server, _ jsoniter.Any) {
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

func (n *DiscoveryFile) OnStop() {

}

func (n *DiscoveryFile) List() []facade.IMember {
	var list []facade.IMember
	for _, member := range n.memberList {
		list = append(list, member)
	}
	return list
}

func (n *DiscoveryFile) GetType(nodeId string) (nodeType string, err error) {
	member, found := n.GetMember(nodeId)
	if found == false {
		return "", cherryError.Errorf("nodeId = %s not found.", nodeId)
	}
	return member.GetNodeType(), nil
}

func (n *DiscoveryFile) GetMember(nodeId string) (facade.IMember, bool) {
	for _, member := range n.memberList {
		if member.GetNodeId() == nodeId {
			return member, true
		}
	}

	return nil, false
}

func (n *DiscoveryFile) OnAddMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	n.onAddListener = append(n.onAddListener, listener)
}

func (n *DiscoveryFile) OnRemoveMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	n.onRemoveListener = append(n.onRemoveListener, listener)
}

func (n *DiscoveryFile) AddMember(member facade.IMember) {
	if _, found := n.GetMember(member.GetNodeId()); found {
		cherryLogger.Warnf("duplicate nodeId. [nodeType = %s], [nodeId = %s]",
			member.GetNodeType(), member.GetNodeId())
		return
	}

	defer n.Unlock()
	n.Lock()

	n.memberList = append(n.memberList, &cherryProto.Member{
		NodeId:   member.GetNodeId(),
		NodeType: member.GetNodeType(),
		Address:  member.GetAddress(),
		Settings: member.GetSettings(),
	})

	for _, listener := range n.onAddListener {
		listener(member)
	}

	cherryLogger.Debugf("add new member. [member = %s]", member)
}

func (n *DiscoveryFile) RemoveMember(nodeId string) {
	defer n.Unlock()
	n.Lock()

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
