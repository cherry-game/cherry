package cherryDiscovery

import (
	"fmt"
	"time"

	ctime "github.com/cherry-game/cherry/extend/time"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
)

// DiscoveryMaster master节点模式(master为单节点)
// 先启动一个master节点
// 其他节点启动时Request(cherry.discovery.register)，到master节点注册
// master节点subscribe(cherry.discovery.register)，返回已注册节点列表
// master节点publish(cherry.discovery.addMember)，当前已注册的节点到
// 所有客户端节点subscribe(cherry.discovery.addMember)，接收新节点
// 所有节点subscribe(cherry.discovery.unregister)，退出时注销节点
type DiscoveryMaster struct {
	DiscoveryDefault
	app              cfacade.IApplication
	thisMember       cfacade.IMember
	masterID         string
	prefix           string
	registerSubject  string
	addSubject       string
	removeSubject    string
	heartbeatSubject string
	thisNodeIDBytes  []byte
	thisMemberBytes  []byte
}

func (m *DiscoveryMaster) Name() string {
	return "nats"
}

func (m *DiscoveryMaster) isMaster() bool {
	return m.app.NodeID() == m.masterID
}

func (m *DiscoveryMaster) isClient() bool {
	return m.app.NodeID() != m.masterID
}

func (m *DiscoveryMaster) buildSubject(subject string) string {
	return fmt.Sprintf(subject, m.prefix, m.masterID)
}

func (m *DiscoveryMaster) Load(app cfacade.IApplication) {
	m.DiscoveryDefault.PreInit()
	m.app = app
	m.loadMember()
	m.init()
}

func (m *DiscoveryMaster) loadMember() {
	// Get nats config
	config := cprofile.GetConfig("cluster").GetConfig(m.Name())
	if config.LastError() != nil {
		clog.Fatalf("[loadMember] Nats config not found. err = %v", config.LastError())
	}

	m.prefix = config.GetString("prefix", "node")

	// Get master node id
	m.masterID = config.GetString("master_node_id")
	if m.masterID == "" {
		clog.Fatal("[loadMember] Master node id not in config.")
	}

	// Default timeout is 3 seconds
	clusterHeartbeatTimeout := m.app.Settings().GetDuration("cluster_heartbeat_timeout", 3) * time.Second

	m.thisMember = &cproto.Member{
		NodeID:           m.app.NodeID(),
		NodeType:         m.app.NodeType(),
		Address:          m.app.RpcAddress(),
		LastAt:           ctime.Now().ToMillisecond(),
		HeartbeatTimeout: clusterHeartbeatTimeout.Milliseconds(),
		//Settings: make(map[string]string),
	}

	if err := m.preloadMarshal(); err != nil {
		clog.Fatalf("[init] Marshal data error. err = %v", err)
	}
}

func (m *DiscoveryMaster) init() {
	m.registerSubject = m.buildSubject("cherry.%s.discovery.%s.register")
	m.addSubject = m.buildSubject("cherry.%s.discovery.%s.add")
	m.removeSubject = m.buildSubject("cherry.%s.discovery.%s.remove")
	m.heartbeatSubject = m.buildSubject("cherry.%s.discovery.%s.heartbeat")

	// Node init
	m.masterInit()
	m.clientInit()

	clog.Infof("[init] Discovery = %s is running.", m.Name())
}

func (m *DiscoveryMaster) masterInit() {
	if m.isMaster() {
		// add master member
		m.AddMember(m.thisMember)
		// subscribe register message
		m.registerSubscribe()
		// subscribe remove message
		m.removeSubscribe()
		// subscribe heartbeat message
		m.heartbeatSubscribe()
	}
}

func (m *DiscoveryMaster) clientInit() {
	if m.isClient() {
		// receive registered node
		m.addSubscribe()
		// subscribe remove message
		m.removeSubscribe()
		// send register&heartbeat message to master node
		go m.send2Master()
	}
}

func (m *DiscoveryMaster) addSubscribe() {
	m.subscribe(m.addSubject, func(msg *nats.Msg) {
		addMember, err := m.bytes2Member(msg.Data)
		if err != nil {
			clog.Warnf("[addSubscribe] bytes to Member error. err = %s", err)
			return
		}

		if _, ok := m.GetMember(addMember.NodeID); !ok {
			m.AddMember(addMember)
		}
	})
}

func (m *DiscoveryMaster) send2Master() {
	for {
		if _, found := m.GetMember(m.masterID); !found {
			m.sendRegister2Master()
		} else {
			m.sendHeartbeat2Master()
		}

		time.Sleep(cnats.ReconnectDelay())
	}
}

func (m *DiscoveryMaster) Stop() {
	if m.isClient() {
		m.sendRemove(m.thisNodeIDBytes)
	}

	clog.Debugf("[Stop] NodeID = %s is unregister", m.app.NodeID())
}

func (m *DiscoveryMaster) sendRemove(data []byte) {
	err := cnats.GetConnect().Publish(m.removeSubject, data)
	if err != nil {
		clog.Warnf("[sendRemove] Publish fail. err = %s", err)
		return
	}
}

func (m *DiscoveryMaster) sendRegister2Master() {
	// register current node to master
	rspData, err := cnats.GetConnect().Request(m.registerSubject, m.thisMemberBytes)
	if err != nil {
		clog.Warnf("[sendRegister2Master] Fail. master = %s, err = %s",
			m.masterID,
			err,
		)
		return
	}

	clog.Infof("[sendRegister2Master] OK. master = %s, member = %s", m.masterID, m.thisMember)

	memberList, err := m.bytes2MemberList(rspData)
	if err != nil {
		clog.Warnf("[sendRegister2Master] Rsp data error. err = %s", err)
		return
	}

	for _, member := range memberList.GetList() {
		if member.GetNodeID() != m.thisMember.GetNodeID() {
			m.AddMember(member)
		}
	}
}

func (m *DiscoveryMaster) sendHeartbeat2Master() {
	err := cnats.GetConnect().Publish(m.heartbeatSubject, m.thisNodeIDBytes)
	if err != nil {
		clog.Warnf("[sendHeartbeat2Master] Publish fail. err = %s", err)
		return
	}
}

func (m *DiscoveryMaster) heartbeatSubscribe() {
	// check heartbeat
	go m.heartbeatCheck()

	m.subscribe(m.heartbeatSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[heartbeatSubscribe] bytes to NodeID error. err = %v", err)
			return
		}

		if value, found := m.GetMember(nodeID); found {
			if protoMember, ok := value.(*cproto.Member); ok {
				// update last heartbeat time
				protoMember.LastAt = ctime.Now().ToMillisecond()
			}
		}
	})
}

func (m *DiscoveryMaster) heartbeatCheck() {
	for {
		m.memberMap.Range(func(key, value any) bool {
			protoMember, ok := value.(*cproto.Member)
			if !ok {
				clog.Warnf("[heartbeatCheck] Member type error. Member = %v", value)
				return true
			}

			//  Determine whether the heartbeat of the node is timed out
			if protoMember.IsTimeout(ctime.Now().NowDiffMillisecond()) {
				m.RemoveMember(protoMember.NodeID)

				nodeIDBytes, err := m.NodeID2Bytes(protoMember.NodeID)
				if err != nil {
					clog.Warnf("[heartbeatCheck] NodeID2Bytes error. err = %s", err)
					return true
				}

				// send nodeID bytes remove message to subscribe nodes
				m.sendRemove(nodeIDBytes)
			}

			return true
		})

		time.Sleep(time.Second) // sleep 1 second
	}
}

func (m *DiscoveryMaster) registerSubscribe() {
	m.subscribe(m.registerSubject, func(msg *nats.Msg) {
		newMember, err := m.bytes2Member(msg.Data)
		if err != nil {
			clog.Warnf("[registerSubscribe] bytes to IMember Unmarshal error. err = %s", err)
			return
		}

		// addMember new member
		m.AddMember(newMember)

		// Response member list to new member node
		memberListBytes, err := m.memberList2Bytes()
		if err != nil {
			clog.Warnf("[registerSubscribe] Marshal fail. err = %s", err)
			return
		}

		err = msg.Respond(memberListBytes)
		if err != nil {
			clog.Warnf("[registerSubscribe] Respond fail. err = %s", err)
			return
		}

		// Publish new member data to all client node
		err = cnats.GetConnect().Publish(m.addSubject, msg.Data)
		if err != nil {
			clog.Warnf("[registerSubscribe] Add subject publish fail. err = %s", err)
			return
		}
	})
}

func (m *DiscoveryMaster) removeSubscribe() {
	m.subscribe(m.removeSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[removeSubscribe] bytes to NodeID error. err = %s", err)
			return
		}

		if nodeID == m.app.NodeID() {
			return
		}
		// remove member
		m.RemoveMember(nodeID)
	})
}

func (m *DiscoveryMaster) preloadMarshal() error {
	var err error
	m.thisNodeIDBytes, err = m.NodeID2Bytes(m.app.NodeID())
	if err != nil {
		return err
	}

	m.thisMemberBytes, err = m.member2Bytes(m.thisMember)
	if err != nil {
		return err
	}

	return nil
}

func (m *DiscoveryMaster) member2Bytes(member cfacade.IMember) ([]byte, error) {
	return m.app.Serializer().Marshal(member)
}

func (m *DiscoveryMaster) memberList2Bytes() ([]byte, error) {
	memberList := &cproto.MemberList{}
	m.memberMap.Range(func(key, value any) bool {
		protoMember := value.(*cproto.Member)
		memberList.List = append(memberList.List, protoMember)
		return true
	})

	return m.app.Serializer().Marshal(memberList)
}

func (m *DiscoveryMaster) bytes2MemberList(data []byte) (*cproto.MemberList, error) {
	memberList := &cproto.MemberList{}
	err := m.app.Serializer().Unmarshal(data, memberList)
	if err != nil {
		return nil, err
	}

	return memberList, nil
}

func (m *DiscoveryMaster) NodeID2Bytes(nodeID string) ([]byte, error) {
	nodeProto := &cproto.NodeID{
		Value: nodeID,
	}

	return m.app.Serializer().Marshal(nodeProto)
}

func (m *DiscoveryMaster) bytes2Member(data []byte) (*cproto.Member, error) {
	member := &cproto.Member{}
	err := m.app.Serializer().Unmarshal(data, member)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (m *DiscoveryMaster) bytes2NodeID(data []byte) (string, error) {
	nodeIDProto := &cproto.NodeID{}
	err := m.app.Serializer().Unmarshal(data, nodeIDProto)
	if err != nil {
		return "", err
	}

	return nodeIDProto.Value, nil
}

func (m *DiscoveryMaster) subscribe(subject string, cb nats.MsgHandler) {
	err := cnats.GetConnect().Subscribe(subject, cb)
	if err != nil {
		clog.Warnf("[subscribe] fail. subject = %s, err = %s", subject, err)
		return
	}
}
