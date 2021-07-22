package cherryCluster

import (
	"context"
	cherryError "github.com/cherry-game/cherry/error"
	cherryGRPC "github.com/cherry-game/cherry/extend/grpc"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc"
	"time"
)

// DiscoveryMaster master类型节点发现
//
// 通过master节点同步成员数据(master为单点服务)
type DiscoveryMaster struct {
	DiscoveryNode
	facade.IApplication
	masterMember  facade.IMember
	rpcServer     *grpc.Server
	clientPool    *connPool
	retryInterval time.Duration
}

func (m *DiscoveryMaster) Name() string {
	return "master"
}

func (m *DiscoveryMaster) Init(app facade.IApplication, rpcServer *grpc.Server, discoveryConfig jsoniter.Any) {
	m.IApplication = app
	m.rpcServer = rpcServer

	// TODO read from profile config
	m.retryInterval = 2 * time.Second

	masterId := discoveryConfig.Get(m.Name()).ToString()
	if masterId == "" {
		panic("master node id not config.")
	}

	// load master node config
	masterNode, err := cherryProfile.LoadNode(masterId)
	if err != nil {
		panic(err)
	}

	m.masterMember = &cherryProto.Member{
		NodeId:   masterNode.NodeId(),
		NodeType: masterNode.NodeType(),
		Address:  masterNode.RpcAddress(),
		Settings: make(map[string]string),
	}

	m.clientPool = newPool(GrpcOptions...) // TODO  GrpcOptions config

	if m.isMaster() {
		m.serverInit()
	} else {
		m.clientInit()
	}
}

func (m *DiscoveryMaster) serverInit() {
	cherryProto.RegisterMasterServiceServer(m.rpcServer, m)

	m.AddMember(m.masterMember)

	cherryLogger.Infof("[discovery = %s] master server node is running. [address = %s]",
		m.Name(), m.masterMember.GetAddress())
}

func (m *DiscoveryMaster) clientInit() {
	m.clientPool.add(m.masterMember)

	client, found := m.clientPool.GetMasterClient(m.masterMember.GetNodeId())
	if found == false {
		cherryLogger.Warnf("get master client not found. [nodeId = %s]", m.masterMember.GetNodeId())
		return
	}

	registerMember := &cherryProto.Member{
		NodeId:   m.NodeId(),
		NodeType: m.NodeType(),
		Address:  m.RpcAddress(),
		Settings: make(map[string]string),
	}

	for {
		rsp, err := client.Register(context.Background(), registerMember)
		if err == nil {
			for _, syncMember := range rsp.List {
				m.AddMember(syncMember)
			}

			cherryLogger.Infof("register to master node success. [masterAddress = %s],[member = %s]",
				m.masterMember.GetAddress(),
				registerMember,
			)
			break
		}

		cherryLogger.Infof("register to master node fail. retry... [error = %s]", err.Error())
		time.Sleep(m.retryInterval)
	}
}

func (m *DiscoveryMaster) OnStop() {
	if m.isMaster() == false {
		client, found := m.clientPool.GetMasterClient(m.masterMember.GetNodeId())
		if found == false {
			cherryLogger.Warnf("get master client not found. [nodeId = %s]", m.masterMember.GetNodeId())
			return
		}

		_, err := client.Unregister(context.Background(), &cherryProto.NodeId{Id: m.NodeId()})
		if err != nil {
			cherryLogger.Warn(err)
			return
		}

		cherryLogger.Debugf("[nodeId = %s] unregister node. [masterAddress = %s]",
			m.NodeId(),
			m.masterMember.GetAddress(),
		)
	}

	m.clientPool.close()
}

func (m *DiscoveryMaster) isMaster() bool {
	return m.NodeId() == m.masterMember.GetNodeId()
}

func (m *DiscoveryMaster) Register(ctx context.Context, newMember *cherryProto.Member) (*cherryProto.MemberList, error) {
	if newMember == nil {
		return nil, cherryError.Errorf("node info is nil [requestIP = %s]", cherryGRPC.GetIP(ctx))
	}

	rspMemberList := &cherryProto.MemberList{}

	for _, member := range m.memberList {
		if member.GetNodeId() == newMember.GetNodeId() {
			continue
		}

		rspMemberList.List = append(rspMemberList.List, member)

		if member.GetNodeId() == m.NodeId() {
			continue
		}

		client, found := m.clientPool.GetMemberClient(member.NodeId)
		if found == false {
			cherryLogger.Warnf("get member client error. [address = %s]", member.GetAddress())
			continue
		}

		_, err := client.NewMember(context.Background(), newMember)
		if err != nil {
			cherryLogger.Warnf("call NewMember() error. [member = %s]", member)
			continue
		}
	}

	// add register member
	m.AddMember(newMember)

	return rspMemberList, nil
}

func (m *DiscoveryMaster) Unregister(ctx context.Context, nodeId *cherryProto.NodeId) (*cherryProto.Response, error) {
	if nodeId == nil || nodeId.Id == "" {
		return nil, cherryError.Errorf("node id is nil. [requestIP = %s]", cherryGRPC.GetIP(ctx))
	}

	// remove member
	m.RemoveMember(nodeId.Id)

	for _, member := range m.memberList {
		if member.GetNodeId() == m.NodeId() {
			continue
		}

		client, found := m.clientPool.GetMemberClient(member.GetNodeId())
		if found == false {
			cherryLogger.Warnf("get member client error. [nodeId = %s]", nodeId)
			continue
		}

		_, err := client.RemoveMember(context.Background(), nodeId)
		if err != nil {
			cherryLogger.Warnf("remove nodeId = %s, error = %s", nodeId, err)
			continue
		}
	}

	return &cherryProto.Response{}, nil
}

func (m *DiscoveryMaster) AddMember(member facade.IMember) {
	m.DiscoveryNode.AddMember(member)
	m.clientPool.add(member)
}

func (m *DiscoveryMaster) RemoveMember(nodeId string) {
	m.DiscoveryNode.RemoveMember(nodeId)
	// clean conn
	m.clientPool.remove(nodeId)

}
