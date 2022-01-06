package cherryDiscovery

import (
	"fmt"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryNats "github.com/cherry-game/cherry/net/cluster/nats"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"time"
)

// DiscoveryNATS master节点模式(master为单节点)
// 先启动一个master节点
// 其他节点启动时Request(cherry.discovery.register)，到master节点注册
// master节点subscribe(cherry.discovery.register)，返回已注册节点列表
// master节点publish(cherry.discovery.add)，当前已注册的节点到
// 所有客户端节点subscribe(cherry.discovery.add)，接收新节点
// 所有节点subscribe(cherry.discovery.unregister)，退出时注销节点
type DiscoveryNATS struct {
	DiscoveryDefault
	facade.IApplication
	masterMember      facade.IMember
	registerSubject   string
	unregisterSubject string
	addSubject        string
}

func (m *DiscoveryNATS) Name() string {
	return "nats"
}

func (m *DiscoveryNATS) Init(app facade.IApplication) {
	m.IApplication = app

	//get nats config
	natsConfig := cherryProfile.Config().Get("cluster").Get(m.Name())
	if natsConfig.LastError() != nil {
		cherryLogger.Fatalf("nats config parameter not found. err = %v", natsConfig.LastError())
	}

	// get master node id
	masterId := natsConfig.Get("master_node_id").ToString()
	if masterId == "" {
		cherryLogger.Fatal("master node id not in config.")
	}

	// load master node config
	masterNode, err := cherryProfile.LoadNode(masterId)
	if err != nil {
		cherryLogger.Fatal(err)
	}

	m.masterMember = &cherryProto.Member{
		NodeId:   masterNode.NodeId(),
		NodeType: masterNode.NodeType(),
		Address:  masterNode.RpcAddress(),
		Settings: make(map[string]string),
	}

	m.init()
}

func (m *DiscoveryNATS) isMaster() bool {
	return m.NodeId() == m.masterMember.GetNodeId()
}

func (m *DiscoveryNATS) isClient() bool {
	return m.NodeId() != m.masterMember.GetNodeId()
}

func (m *DiscoveryNATS) init() {
	masterNodeId := m.masterMember.GetNodeId()
	m.registerSubject = fmt.Sprintf("cherry.discovery.%s.register", masterNodeId)
	m.unregisterSubject = fmt.Sprintf("cherry.discovery.%s.unregister", masterNodeId)
	m.addSubject = fmt.Sprintf("cherry.discovery.%s.add", masterNodeId)

	m.subscribe(m.unregisterSubject, func(msg *nats.Msg) {
		unregisterMember := &cherryProto.Member{}
		err := m.Unmarshal(msg.Data, unregisterMember)
		if err != nil {
			cherryLogger.Warnf("err = %s", err)
			return
		}

		if unregisterMember.NodeId == m.NodeId() {
			return
		}

		// remove member
		m.RemoveMember(unregisterMember.NodeId)
	})

	m.serverInit()
	m.clientInit()

	cherryLogger.Infof("[discovery = %s] is running.", m.Name())
}

func (m *DiscoveryNATS) serverInit() {
	if m.isMaster() == false {
		return
	}

	//add master node
	m.AddMember(m.masterMember)

	// subscribe register message
	m.subscribe(m.registerSubject, func(msg *nats.Msg) {
		newMember := &cherryProto.Member{}
		err := m.Unmarshal(msg.Data, newMember)
		if err != nil {
			cherryLogger.Warnf("err = %s", err)
			return
		}

		// add new member
		m.AddMember(newMember)

		// response member list
		rspMemberList := &cherryProto.MemberList{}
		for _, member := range m.memberList {
			if member.GetNodeId() == newMember.GetNodeId() {
				continue
			}

			if member.GetNodeId() == m.NodeId() {
				continue
			}

			rspMemberList.List = append(rspMemberList.List, member)
		}

		rspData, err := m.Marshal(rspMemberList)
		if err != nil {
			cherryLogger.Warnf("marshal fail. err = %s", err)
			return
		}

		// response member list
		err = msg.Respond(rspData)
		if err != nil {
			cherryLogger.Warnf("respond fail. err = %s", err)
			return
		}

		// publish add new node
		err = cherryNats.Conn().Publish(m.addSubject, msg.Data)
		if err != nil {
			cherryLogger.Warnf("publish fail. err = %s", err)
			return
		}
	})
}

func (m *DiscoveryNATS) clientInit() {
	if m.isClient() == false {
		return
	}

	registerMember := &cherryProto.Member{
		NodeId:   m.NodeId(),
		NodeType: m.NodeType(),
		Address:  m.RpcAddress(),
		Settings: make(map[string]string),
	}

	bytesData, err := m.Marshal(registerMember)
	if err != nil {
		cherryLogger.Warnf("err = %s", err)
		return
	}

	// receive registered node
	m.subscribe(m.addSubject, func(msg *nats.Msg) {
		addMember := &cherryProto.Member{}
		err := m.Unmarshal(msg.Data, addMember)
		if err != nil {
			cherryLogger.Warnf("err = %s", err)
			return
		}

		if _, ok := m.GetMember(addMember.NodeId); ok == false {
			m.AddMember(addMember)
		}
	})

	for {
		// register current node to master
		rsp, err := cherryNats.Conn().Request(m.registerSubject, bytesData, cherryNats.Conn().RequestTimeout)
		if err != nil {
			cherryLogger.Warnf("register node to [master = %s] fail. [address = %s] [err = %s]",
				m.masterMember.GetNodeId(),
				cherryNats.Conn().Address,
				err,
			)
			time.Sleep(cherryNats.Conn().RequestTimeout)
			continue
		}

		cherryLogger.Infof("register node to [master = %s]. [member = %s]",
			m.masterMember.GetNodeId(),
			registerMember,
		)

		memberList := cherryProto.MemberList{}
		err = m.Unmarshal(rsp.Data, &memberList)
		if err != nil {
			cherryLogger.Warnf("err = %s", err)
			time.Sleep(cherryNats.Conn().RequestTimeout)
			continue
		}

		for _, member := range memberList.GetList() {
			m.AddMember(member)
		}

		break
	}
}

func (m *DiscoveryNATS) OnStop() {
	if m.isClient() {
		thisMember := &cherryProto.Member{
			NodeId: m.NodeId(),
		}

		bytesData, err := m.Marshal(thisMember)
		if err != nil {
			cherryLogger.Warnf("marshal fail. err = %s", err)
			return
		}

		err = cherryNats.Conn().Publish(m.unregisterSubject, bytesData)
		if err != nil {
			cherryLogger.Warnf("publish fail. err = %s", err)
			return
		}

		cherryLogger.Debugf("[nodeId = %s] unregister node to [[master = %s]",
			m.NodeId(),
			m.masterMember.GetNodeId(),
		)
	}

	cherryNats.Conn().Close()
}

func (m *DiscoveryNATS) subscribe(subject string, cb nats.MsgHandler) {
	_, err := cherryNats.Conn().Subscribe(subject, cb)
	if err != nil {
		cherryLogger.Warnf("subscribe fail. err = %s", err)
		return
	}
}
