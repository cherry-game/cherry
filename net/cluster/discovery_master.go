package cherryCluster

import (
	"context"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryError "github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	jsoniter "github.com/json-iterator/go"
	"sync"
)

// DiscoveryMaster master类型节点发现
//
// 通过master进程服务同步节点数据(master为单点服务)
type DiscoveryMaster struct {
	sync.RWMutex
	DiscoveryNode
	app              facade.IApplication
	component        *Component
	masterNode       facade.INode
	onAddListener    []facade.MemberListener
	onRemoveListener []facade.MemberListener
}

func (m *DiscoveryMaster) Name() string {
	return "master"
}

func (m *DiscoveryMaster) Init(app facade.IApplication, discoveryConfig jsoniter.Any) {
	m.app = app
	m.component = app.Find(cherryConst.ClusterComponent).(*Component)
	m.DiscoveryNode.Init(app, discoveryConfig)

	masterId := discoveryConfig.Get(m.Name()).ToString()
	if masterId == "" {
		panic("master node id not config.")
	}

	if m.app.IsMaster() == false {
		// 获取master 节点信息
		// 连接 master服务器
		// 注册当前节点信息到master服务器

		var err error
		m.masterNode, err = cherryProfile.LoadNode(masterId)
		if err != nil {
			panic(err)
		}

	}

}

func (m *DiscoveryMaster) OnAddMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	m.onAddListener = append(m.onAddListener, listener)
}

func (m *DiscoveryMaster) OnRemoveMember(listener facade.MemberListener) {
	if listener == nil {
		return
	}
	m.onRemoveListener = append(m.onRemoveListener, listener)
}

// Register 向master注册节点
func (m *DiscoveryMaster) Register(_ context.Context, registerMember *cherryProto.Member) (*cherryProto.MemberList, error) {
	if registerMember == nil {
		return nil, cherryError.Error("node info is nil")
	}

	for _, member := range m.memberList {
		if member.GetNodeId() == registerMember.GetNodeId() {
			return nil, cherryError.Errorf("node = %s has registered.", registerMember.GetNodeId())
		}
	}

	pushMember := &cherryProto.MemberList{}
	for _, member := range m.memberList {
		pushMember.List = append(pushMember.List, member)

		if member.GetNodeId() == m.app.NodeId() {
			m.AddMember(registerMember)
			continue
		}

		client, err := m.GetMemberClient(member.Address)
		if err != nil {
			return nil, err
		}

		_, err = client.NewMember(context.Background(), registerMember)
		if err != nil {
			return nil, err
		}
	}

	return pushMember, nil
}

func (m *DiscoveryMaster) GetMemberClient(address string) (cherryProto.MemberServiceClient, error) {
	pool, err := m.component.rpcClient.getConnPool(address)
	if err != nil {
		return nil, err
	}

	return cherryProto.NewMemberServiceClient(pool.Get()), nil
}

// Unregister 向master注销节点
func (m *DiscoveryMaster) Unregister(_ context.Context, nodeId *cherryProto.NodeId) (*cherryProto.Response, error) {
	_, found := m.Get(nodeId.Id)
	if found == false {
		return nil, cherryError.Errorf("nodeId = %d not found.", nodeId.Id)
	}

	for _, member := range m.memberList {
		if member.GetNodeId() == m.app.NodeId() {
			m.RemoveMember(nodeId.Id)
			continue
		}

		client, err := m.GetMemberClient(member.Address)
		if err != nil {
			return nil, err
		}

		_, err = client.RemoveMember(context.Background(), nodeId)
		if err != nil {
			return nil, err
		}
	}

	return &cherryProto.Response{}, nil
}

func (m *DiscoveryMaster) AddMember(member *cherryProto.Member) {
	if _, found := m.Get(member.GetNodeId()); found {
		cherryLogger.Errorf("nodeType = %s, nodeId = %s, duplicate nodeId.",
			member.GetNodeType(), member.GetNodeId())
		return
	}

	defer m.Unlock()
	m.Lock()

	m.memberList = append(m.memberList, member)

	for _, listener := range m.onAddListener {
		listener(member)
	}
}

func (m *DiscoveryMaster) RemoveMember(nodeId string) {
	defer m.Unlock()
	m.Lock()

	var member *cherryProto.Member
	for i := 0; i < len(m.memberList); i++ {
		member = m.memberList[i]

		if member.NodeId == nodeId {
			m.memberList = append(m.memberList[:i], m.memberList[i+1:]...)
			break
		}
	}

	for _, listener := range m.onRemoveListener {
		listener(member)
	}
}

//// RequestCloseSession 请求关闭session
//func (c *DiscoveryMaster) RequestCloseSession(_ context.Context, sessionId *cherryProto.SessionId) (*cherryProto.Response, error) {
//    for _, member := range c.memberList {
//        if member.Id == c.App().NodeId() {
//            continue
//        }
//
//        pool, err := c.rpcClient.getConnPool(member.Address)
//        if err != nil {
//            return nil, err
//        }
//
//        // 通知所有节点，删除sessionId
//        client := cherryProto.NewMemberClient(pool.GetDiscovery())
//        _, err = client.CloseSession(context.Background(), sessionId)
//        if err != nil {
//            return nil, err
//        }
//    }
//
//    return nil, nil
//}
