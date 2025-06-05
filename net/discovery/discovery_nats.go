package cherryDiscovery

import (
	"fmt"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
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
	app               cfacade.IApplication
	thisMember        cfacade.IMember
	thisMemberBytes   []byte
	masterMember      cfacade.IMember
	registerSubject   string
	unregisterSubject string
	addSubject        string
	checkSubject      string
}

func (m *DiscoveryNATS) Name() string {
	return "nats"
}

func (m *DiscoveryNATS) isMaster() bool {
	return m.app.NodeID() == m.masterMember.GetNodeID()
}

func (m *DiscoveryNATS) isClient() bool {
	return m.app.NodeID() != m.masterMember.GetNodeID()
}

func (m *DiscoveryNATS) buildSubject(subject string) string {
	return fmt.Sprintf(subject, m.masterMember.GetNodeID())
}

func (m *DiscoveryNATS) Load(app cfacade.IApplication) {
	m.DiscoveryDefault.PreInit()
	m.app = app
	m.loadMember()
	m.init()
}

func (m *DiscoveryNATS) loadMember() {
	m.thisMember = &cproto.Member{
		NodeID:   m.app.NodeID(),
		NodeType: m.app.NodeType(),
		Address:  m.app.RpcAddress(),
		Settings: make(map[string]string),
	}

	memberBytes, err := m.app.Serializer().Marshal(m.thisMember)
	if err != nil {
		clog.Warnf("err = %s", err)
		return
	}

	m.thisMemberBytes = memberBytes

	//get nats config
	config := cprofile.GetConfig("cluster").GetConfig(m.Name())
	if config.LastError() != nil {
		clog.Fatalf("nats config parameter not found. err = %v", config.LastError())
	}

	// get master node id
	masterID := config.GetString("master_node_id")
	if masterID == "" {
		clog.Fatal("master node id not in config.")
	}

	// load master node config
	masterNode, err := cprofile.LoadNode(masterID)
	if err != nil {
		clog.Fatal(err)
	}

	m.masterMember = &cproto.Member{
		NodeID:   masterNode.NodeID(),
		NodeType: masterNode.NodeType(),
		Address:  masterNode.RpcAddress(),
		Settings: make(map[string]string),
	}
}

func (m *DiscoveryNATS) init() {
	m.registerSubject = m.buildSubject("cherry.discovery.%s.register")
	m.unregisterSubject = m.buildSubject("cherry.discovery.%s.unregister")
	m.addSubject = m.buildSubject("cherry.discovery.%s.addMember")
	m.checkSubject = m.buildSubject("cherry.discovery.%s.check")

	m.subscribe(m.unregisterSubject, func(msg *nats.Msg) {
		unregisterMember := &cproto.Member{}
		err := m.app.Serializer().Unmarshal(msg.Data, unregisterMember)
		if err != nil {
			clog.Warnf("err = %s", err)
			return
		}

		if unregisterMember.NodeID == m.app.NodeID() {
			return
		}

		// remove member
		m.RemoveMember(unregisterMember.NodeID)
	})

	m.serverInit()
	m.clientInit()

	clog.Infof("[discovery = %s] is running.", m.Name())
}

func (m *DiscoveryNATS) serverInit() {
	if !m.isMaster() {
		return
	}

	//addMember master node
	m.AddMember(m.masterMember)

	// subscribe register message
	m.subscribe(m.registerSubject, func(msg *nats.Msg) {
		newMember := &cproto.Member{}
		err := m.app.Serializer().Unmarshal(msg.Data, newMember)
		if err != nil {
			clog.Warnf("IMember Unmarshal[name = %s] error. dataLen = %+v, err = %s",
				m.app.Serializer().Name(),
				len(msg.Data),
				err,
			)
			return
		}

		// addMember new member
		m.AddMember(newMember)

		// response member list
		memberList := &cproto.MemberList{}

		m.memberMap.Range(func(key, value any) bool {
			protoMember := value.(*cproto.Member)
			if protoMember.NodeID != newMember.NodeID {
				memberList.List = append(memberList.List, protoMember)
			}

			return true
		})

		rspData, err := m.app.Serializer().Marshal(memberList)
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
		err = cnats.Get().Publish(m.addSubject, msg.Data)
		if err != nil {
			clog.Warnf("publish fail. err = %s", err)
			return
		}
	})

	// subscribe check message
	m.subscribe(m.checkSubject, func(msg *nats.Msg) {
		msg.Respond(nil)
	})
}

func (m *DiscoveryNATS) clientInit() {
	if !m.isClient() {
		return
	}

	// receive registered node
	m.subscribe(m.addSubject, func(msg *nats.Msg) {
		addMember := &cproto.Member{}
		err := m.app.Serializer().Unmarshal(msg.Data, addMember)
		if err != nil {
			clog.Warnf("err = %s", err)
			return
		}

		if _, ok := m.GetMember(addMember.NodeID); !ok {
			m.AddMember(addMember)
		}
	})

	go m.checkMaster()
}

func (m *DiscoveryNATS) checkMaster() {
	for {
		_, found := m.GetMember(m.masterMember.GetNodeID())
		if !found {
			m.registerToMaster()
		}

		time.Sleep(cnats.Get().ReconnectDelay())
	}
}

func (m *DiscoveryNATS) registerToMaster() {
	// register current node to master
	rsp, err := cnats.Get().Request(m.registerSubject, m.thisMemberBytes)
	if err != nil {
		clog.Warnf("register node to [master = %s] fail. [address = %s] [err = %s]",
			m.masterMember.GetNodeID(),
			cnats.Get().Address(),
			err,
		)
		return
	}

	clog.Infof("register node to [master = %s]. [member = %s]",
		m.masterMember.GetNodeID(),
		m.thisMember,
	)

	memberList := cproto.MemberList{}
	err = m.app.Serializer().Unmarshal(rsp.Data, &memberList)
	if err != nil {
		clog.Warnf("err = %s", err)
		return
	}

	for _, member := range memberList.GetList() {
		m.AddMember(member)
	}
}

func (m *DiscoveryNATS) Stop() {
	err := cnats.Get().Publish(m.unregisterSubject, m.thisMemberBytes)
	if err != nil {
		clog.Warnf("publish fail. err = %s", err)
		return
	}

	clog.Debugf("[nodeID = %s] unregister node to [master = %s]",
		m.app.NodeID(),
		m.masterMember.GetNodeID(),
	)
}

func (m *DiscoveryNATS) subscribe(subject string, cb nats.MsgHandler) {
	_, err := cnats.Get().Subscribe(subject, cb)
	if err != nil {
		clog.Warnf("subscribe fail. err = %s", err)
		return
	}
}
