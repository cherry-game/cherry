package cherryDiscovery

import (
	"fmt"
	cherryNATS "github.com/cherry-game/cherry/extend/nats"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"time"
)

// DiscoveryNats master节点模式(master为单点)
type DiscoveryNats struct {
	DiscoveryDefault
	facade.IApplication
	masterMember      facade.IMember
	nats              *nats.Conn
	natsConfig        *cherryProfile.NatsConfig
	registerSubject   string
	unregisterSubject string
	addSubject        string
}

func (m *DiscoveryNats) Name() string {
	return "nats"
}

func (m *DiscoveryNats) Init(app facade.IApplication, discoveryConfig jsoniter.Any, _ ...interface{}) {
	m.IApplication = app

	//get master config
	masterConfig := discoveryConfig.Get(m.Name())
	masterId := masterConfig.Get("master_node_id").ToString()
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

	m.natsConfig = cherryProfile.NewNatsConfig()

	m.nats, err = cherryNATS.Connect(m.natsConfig)
	if err != nil {
		cherryLogger.Warnf("err = %s", err)
	}

	m.init()
}

func (m *DiscoveryNats) isMaster() bool {
	return m.NodeId() == m.masterMember.GetNodeId()
}

func (m *DiscoveryNats) isClient() bool {
	return m.NodeId() != m.masterMember.GetNodeId()
}

func (m *DiscoveryNats) init() {
	masterNodeId := m.masterMember.GetNodeId()
	m.registerSubject = fmt.Sprintf("cherry.discovery.%s.register", masterNodeId)
	m.unregisterSubject = fmt.Sprintf("cherry.discovery.%s.unregister", masterNodeId)
	m.addSubject = fmt.Sprintf("cherry.discovery.%s.add", masterNodeId)

	//所有客户端节点启动时发送cherry.discovery.register
	//master节点订阅cherry.discovery.register，返回已注册节点列表
	//同时master节点会publish当前已注册的节点到cherry.discovery.add
	//所有客户端节点订阅 cherry.discovery.add，接收新节点
	//所有节点订阅 cherry.discovery.unregister，注销节点

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

	cherryLogger.Infof("[discoveryName = %s] is running.", m.Name())

	m.serverInit()
	m.clientInit()
}

func (m *DiscoveryNats) serverInit() {
	if m.isMaster() == false {
		return
	}

	//add master node
	m.AddMember(m.masterMember)

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

		err = msg.Respond(rspData)
		if err != nil {
			cherryLogger.Warnf("respond fail. err = %s", err)
			return
		}

		// publish add node
		err = m.nats.Publish(m.addSubject, msg.Data)
		if err != nil {
			cherryLogger.Warnf("publish fail. err = %s", err)
			return
		}
	})
}

func (m *DiscoveryNats) clientInit() {
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

	for {
		rsp, err := m.nats.Request(m.registerSubject, bytesData, m.natsConfig.RequestTimeout)
		if err != nil {
			cherryLogger.Warnf("register node to [master = %s] fail. [address = %s] [err = %s]",
				m.masterMember.GetNodeId(),
				m.natsConfig.Address,
				err,
			)
			time.Sleep(m.natsConfig.RequestTimeout)
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
			time.Sleep(m.natsConfig.RequestTimeout)
			continue
		}

		for _, member := range memberList.GetList() {
			m.AddMember(member)
		}

		break
	}

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
}

func (m *DiscoveryNats) OnStop() {
	if m.isClient() {
		thisMember := &cherryProto.Member{
			NodeId: m.NodeId(),
		}

		bytesData, err := m.Marshal(thisMember)
		if err != nil {
			cherryLogger.Warnf("marshal fail. err = %s", err)
			return
		}

		err = m.nats.Publish(m.unregisterSubject, bytesData)
		if err != nil {
			cherryLogger.Warnf("publish fail. err = %s", err)
			return
		}

		cherryLogger.Debugf("[nodeId = %s] unregister node to [[master = %s]",
			m.NodeId(),
			m.masterMember.GetNodeId(),
		)
	}

	m.nats.Close()
}

func (m *DiscoveryNats) subscribe(subject string, cb nats.MsgHandler) {
	_, err := m.nats.Subscribe(subject, cb)
	if err != nil {
		cherryLogger.Warnf("subscribe fail. err = %s", err)
		return
	}
}
