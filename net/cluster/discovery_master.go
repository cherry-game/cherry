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
	masterNode    facade.INode
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
	m.masterNode = masterNode

	m.clientPool = newPool(1, GrpcOptions...) // TODO  GrpcOptions config

	m.serverInit()
	m.clientInit()
}

func (m *DiscoveryMaster) serverInit() {
	if m.isMaster() == false {
		return
	}

	cherryProto.RegisterMasterServiceServer(m.rpcServer, m)

	m.AddMember(&cherryProto.Member{
		NodeId:   m.masterNode.NodeId(),
		NodeType: m.masterNode.NodeType(),
		Address:  m.masterNode.RpcAddress(),
		Settings: make(map[string]string),
	})

	cherryLogger.Infof("[discovery = %s] master server node is running. [address = %s]",
		m.Name(), m.masterNode.RpcAddress())
}

func (m *DiscoveryMaster) clientInit() {
	if m.isMaster() {
		return
	}

	client, err := m.clientPool.GetMasterClient(m.masterNode.RpcAddress())
	if err != nil {
		panic(err)
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

			cherryLogger.Infof("this member register to master server node. [masterAddress = %s],[member = %s]",
				m.masterNode.RpcAddress(),
				registerMember,
			)
			break
		}

		cherryLogger.Infof("this member register to master server node fail. retry... [error = %s]", err)
		time.Sleep(m.retryInterval)
	}
}

func (m *DiscoveryMaster) OnStop() {
	if m.isMaster() == false {
		client, err := m.clientPool.GetMasterClient(m.masterNode.RpcAddress())
		if err != nil {
			cherryLogger.Warn(err)
			return
		}

		_, err = client.Unregister(context.Background(), &cherryProto.NodeId{Id: m.NodeId()})
		if err != nil {
			cherryLogger.Warn(err)
			return
		}

		cherryLogger.Debugf("[nodeId = %s] unregister node. [masterAddress = %s]",
			m.NodeId(),
			m.masterNode.RpcAddress(),
		)
	}

	m.clientPool.close()
}

func (m *DiscoveryMaster) isMaster() bool {
	return m.NodeId() == m.masterNode.NodeId()
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

		if member.GetNodeId() == m.NodeId() || member.GetNodeId() == newMember.GetNodeId() {
			continue
		}

		client, err := m.clientPool.GetMemberClient(member.GetAddress())
		if err != nil {
			cherryLogger.Warnf("get member client error. [address = %s]", member.GetAddress())
			continue
		}

		_, err = client.NewMember(context.Background(), newMember)
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

		client, err := m.clientPool.GetMemberClient(member.GetAddress())
		if err != nil {
			cherryLogger.Warnf("get member client error. [nodeId = %s], [error = %s]", nodeId, err)
			return nil, err
		}

		_, err = client.RemoveMember(context.Background(), nodeId)
		if err != nil {
			cherryLogger.Warnf("remove nodeId = %s, error = %s", nodeId, err)
			return nil, err
		}
	}

	return &cherryProto.Response{}, nil
}
