package cherryDiscovery

import (
	"fmt"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/cluster/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"time"
)

// DiscoveryNATS master节点模式(master为单节点)
// 先启动一个master节点
// 其他节点启动时Request(cherry.discovery.register)，到master节点注册
// master节点subscribe(cherry.discovery.register)，返回已注册节点列表
// master节点publish(cherry.discovery.addMember)，当前已注册的节点到
// 所有客户端节点subscribe(cherry.discovery.addMember)，接收新节点
// 所有节点subscribe(cherry.discovery.unregister)，退出时注销节点
type DiscoveryNATS struct {
	DiscoveryDefault
	cfacade.IApplication
	masterMember      cfacade.IMember
	registerSubject   string
	unregisterSubject string
	addSubject        string
}

func (m *DiscoveryNATS) Name() string {
	return "nats"
}

func (m *DiscoveryNATS) Init(app cfacade.IApplication) {
	m.IApplication = app

	//get nats config
	clusterConfig := cprofile.GetConfig("cluster").GetConfig(m.Name())
	if clusterConfig.LastError() != nil {
		clog.Fatalf("nats config parameter not found. err = %v", clusterConfig.LastError())
	}

	// get master node id
	masterId := clusterConfig.GetString("master_node_id")
	if masterId == "" {
		clog.Fatal("master node id not in config.")
	}

	// load master node config
	masterNode, err := cprofile.LoadNode(masterId)
	if err != nil {
		clog.Fatal(err)
	}

	m.masterMember = &cproto.Member{
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
	m.addSubject = fmt.Sprintf("cherry.discovery.%s.addMember", masterNodeId)

	m.subscribe(m.unregisterSubject, func(msg *nats.Msg) {
		unregisterMember := &cproto.Member{}
		err := m.Unmarshal(msg.Data, unregisterMember)
		if err != nil {
			clog.Warnf("err = %s", err)
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

	clog.Infof("[discovery = %s] is running.", m.Name())
}

func (m *DiscoveryNATS) serverInit() {
	if m.isMaster() == false {
		return
	}

	//addMember master node
	m.AddMember(m.masterMember)

	// subscribe register message
	m.subscribe(m.registerSubject, func(msg *nats.Msg) {
		newMember := &cproto.Member{}
		err := m.Unmarshal(msg.Data, newMember)
		if err != nil {
			clog.Warnf("data = %+v, err = %s", msg.Data, err)
			return
		}

		// addMember new member
		m.AddMember(newMember)

		// response member list
		rspMemberList := &cproto.MemberList{}
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
			clog.Warnf("marshal fail. err = %s", err)
			return
		}

		// response member list
		err = msg.Respond(rspData)
		if err != nil {
			clog.Warnf("respond fail. err = %s", err)
			return
		}

		// publish addMember new node
		err = cnats.Publish(m.addSubject, msg.Data)
		if err != nil {
			clog.Warnf("publish fail. err = %s", err)
			return
		}
	})
}

func (m *DiscoveryNATS) clientInit() {
	if m.isClient() == false {
		return
	}

	registerMember := &cproto.Member{
		NodeId:   m.NodeId(),
		NodeType: m.NodeType(),
		Address:  m.RpcAddress(),
		Settings: make(map[string]string),
	}

	bytesData, err := m.Marshal(registerMember)
	if err != nil {
		clog.Warnf("err = %s", err)
		return
	}

	// receive registered node
	m.subscribe(m.addSubject, func(msg *nats.Msg) {
		addMember := &cproto.Member{}
		err := m.Unmarshal(msg.Data, addMember)
		if err != nil {
			clog.Warnf("err = %s", err)
			return
		}

		if _, ok := m.GetMember(addMember.NodeId); ok == false {
			m.AddMember(addMember)
		}
	})

	for {
		// register current node to master
		rsp, err := cnats.Request(m.registerSubject, bytesData, cnats.App().RequestTimeout)
		if err != nil {
			clog.Warnf("register node to [master = %s] fail. [address = %s] [err = %s]",
				m.masterMember.GetNodeId(),
				cnats.App().Address,
				err,
			)
			time.Sleep(cnats.App().RequestTimeout)
			continue
		}

		clog.Infof("register node to [master = %s]. [member = %s]",
			m.masterMember.GetNodeId(),
			registerMember,
		)

		memberList := cproto.MemberList{}
		err = m.Unmarshal(rsp.Data, &memberList)
		if err != nil {
			clog.Warnf("err = %s", err)
			time.Sleep(cnats.App().RequestTimeout)
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
		thisMember := &cproto.Member{
			NodeId: m.NodeId(),
		}

		bytesData, err := m.Marshal(thisMember)
		if err != nil {
			clog.Warnf("marshal fail. err = %s", err)
			return
		}

		err = cnats.Publish(m.unregisterSubject, bytesData)
		if err != nil {
			clog.Warnf("publish fail. err = %s", err)
			return
		}

		clog.Debugf("[nodeId = %s] unregister node to [master = %s]",
			m.NodeId(),
			m.masterMember.GetNodeId(),
		)
	}

	cnats.Close()
}

func (m *DiscoveryNATS) subscribe(subject string, cb nats.MsgHandler) {
	_, err := cnats.Subscribe(subject, cb)
	if err != nil {
		clog.Warnf("subscribe fail. err = %s", err)
		return
	}
}
