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

var registerRequired = []byte{0x01}

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
	isStopChan       chan struct{}
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
	m.isStopChan = make(chan struct{})
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
	clusterHeartbeatTimeout := m.app.Settings().GetInt64("cluster_heartbeat_timeout", 3) * ctime.MillisecondsPerSecond

	m.thisMember = &cproto.Member{
		NodeID:           m.app.NodeID(),
		NodeType:         m.app.NodeType(),
		Address:          m.app.RpcAddress(),
		LastAt:           ctime.Now().ToMillisecond(),
		HeartbeatTimeout: clusterHeartbeatTimeout,
		Settings:         make(map[string]string),
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
			if addMember.NodeID != m.thisMember.GetNodeID() {
				m.AddMember(addMember)
			}
		}
	})
}

func (m *DiscoveryMaster) send2Master() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.isStopChan:
			return
		case <-ticker.C:
			needRegister := m.sendHeartbeat2Master()
			if needRegister {
				m.sendRegister2Master()
			}
		}
	}
}

func (m *DiscoveryMaster) sendHeartbeat2Master() bool {
	reqID := cnats.NewStringReqID()
	rspData, err := cnats.RequestSync(reqID, m.heartbeatSubject, m.thisNodeIDBytes)
	if err != nil {
		clog.Warnf("[sendHeartbeat2Master] Fail. master = %s, err = %s", m.masterID, err)
		return false
	}

	return len(rspData) > 0
}

func (m *DiscoveryMaster) sendRegister2Master() {
	reqID := cnats.NewStringReqID()
	rspData, err := cnats.RequestSync(reqID, m.registerSubject, m.thisMemberBytes)
	if err != nil {
		clog.Warnf("[sendRegister2Master] Fail. master = %s, err = %s", m.masterID, err)
		return
	}

	clog.Infof("[sendRegister2Master] OK. master = %s", m.masterID)

	memberList, err := m.bytes2MemberList(rspData)
	if err != nil {
		clog.Warnf("[sendRegister2Master] Rsp data error. err = %s", err)
		return
	}

	for _, member := range memberList.GetList() {
		if _, ok := m.GetMember(member.GetNodeID()); !ok {
			m.AddMember(member)
		}
	}
}

func (m *DiscoveryMaster) Stop() {
	if m.isClient() {
		m.sendRemove(m.thisNodeIDBytes)
	}

	close(m.isStopChan)

	clog.Debugf("[Stop] NodeID = %s is unregister", m.app.NodeID())
}

func (m *DiscoveryMaster) sendRemove(data []byte) {
	err := cnats.Publish(m.removeSubject, data)
	if err != nil {
		clog.Warnf("[sendRemove] Publish fail. err = %s", err)
	}
}

func (m *DiscoveryMaster) heartbeatCheck() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.isStopChan:
			return
		case <-ticker.C:
			m.checkMemberTimeout()
		}
	}
}

func (m *DiscoveryMaster) checkMemberTimeout() {
	now := ctime.Now().ToMillisecond()
	m.memberMap.Range(func(key, value any) bool {
		protoMember, ok := value.(*cproto.Member)
		if !ok {
			clog.Warnf("[checkMemberTimeout] Member type error. Member = %v", value)
			return true
		}

		if protoMember.NodeID == m.thisMember.GetNodeID() {
			return true
		}

		if protoMember.IsTimeout(now) {
			m.RemoveMember(protoMember.NodeID)

			nodeIDBytes, err := m.NodeID2Bytes(protoMember.NodeID)
			if err != nil {
				clog.Warnf("[checkMemberTimeout] NodeID2Bytes error. err = %s", err)
				return true
			}

			m.sendRemove(nodeIDBytes)
		}

		return true
	})
}

func (m *DiscoveryMaster) registerSubscribe() {
	m.subscribe(m.registerSubject, func(msg *nats.Msg) {
		newMember, err := m.bytes2Member(msg.Data)
		if err != nil {
			clog.Warnf("[registerSubscribe] bytes to Member error. err = %s", err)
			return
		}

		// update last heartbeat time
		newMember.LastAt = ctime.Now().ToMillisecond()
		m.AddMember(newMember)
		m.sendAdd(newMember)
		m.replyMemberList(msg)
	})
}

func (m *DiscoveryMaster) heartbeatSubscribe() {
	go m.heartbeatCheck()

	m.subscribe(m.heartbeatSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[heartbeatSubscribe] bytes to NodeID error. err = %v", err)
			return
		}

		reqID := msg.Header.Get(cnats.REQ_ID)

		if value, found := m.GetMember(nodeID); found {
			// known node: update heartbeat + reply empty
			if protoMember, ok := value.(*cproto.Member); ok {
				protoMember.LastAt = ctime.Now().ToMillisecond()
			}
			cnats.ReplySync(reqID, msg.Reply, nil)
		} else {
			// unknown node: reply marker to trigger full registration
			cnats.ReplySync(reqID, msg.Reply, registerRequired)
		}
	})
}

func (m *DiscoveryMaster) replyMemberList(msg *nats.Msg) {
	memberListBytes, err := m.memberList2Bytes()
	if err != nil {
		clog.Warnf("[replyMemberList] Marshal fail. err = %s", err)
		return
	}

	reqID := msg.Header.Get(cnats.REQ_ID)
	err = cnats.ReplySync(reqID, msg.Reply, memberListBytes)
	if err != nil {
		clog.Warnf("[replyMemberList] Reply fail. err = %s", err)
	}
}

func (m *DiscoveryMaster) sendAdd(member *cproto.Member) {
	memberBytes, err := m.member2Bytes(member)
	if err != nil {
		clog.Warnf("[sendAdd] Marshal fail. err = %s", err)
		return
	}

	err = cnats.Publish(m.addSubject, memberBytes)
	if err != nil {
		clog.Warnf("[sendAdd] Publish fail. err = %s", err)
	}
}

func (m *DiscoveryMaster) removeSubscribe() {
	m.subscribe(m.removeSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[removeSubscribe] bytes to NodeID error. err = %s", err)
			return
		}

		if m.isMaster() && nodeID == m.app.NodeID() {
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
	err := cnats.Subscribe(subject, cb)
	if err != nil {
		clog.Warnf("[subscribe] fail. subject = %s, err = %s", subject, err)
		return
	}
}
