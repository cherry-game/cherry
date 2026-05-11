package cherryDiscovery

import (
	"context"
	"fmt"
	"time"

	ctime "github.com/cherry-game/cherry/extend/time"
	cfacade "github.com/cherry-game/cherry/facade"
	cherryFacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
)

var registerRequired = []byte{0x01}

// ComponentMaster master节点模式(master为单节点)
// 先启动一个master节点
// 其他节点启动时发送注册(cherry.discovery.register)，到master节点
// master节点订阅(cherry.discovery.register)，返回已注册节点列表
// master节点推送(cherry.discovery.addMember)，当前已注册的节点
// 所有客户端节点订阅(cherry.discovery.addMember)，接收新节点
// 所有节点退出时请求(cherry.discovery.remove)，到master节点
type (
	ComponentMaster struct {
		ComponentDefault
		natsSubjects
		thisMember *cproto.Member
		masterID   string
		ctx        context.Context
		cancel     context.CancelFunc
	}

	natsSubjects struct {
		prefix           string
		registerSubject  string
		addSubject       string
		updateSubject    string
		removeSubject    string
		heartbeatSubject string
	}
)

func NewMaster() ComponentMaster {
	return ComponentMaster{}
}

func (m *ComponentMaster) Mode() string {
	return "nats"
}

func (m *ComponentMaster) Init() {
	m.ComponentDefault.InitFields()
	m.init()
}

func (m *ComponentMaster) UpdateSetting(key, value string) {
	m.thisMember.UpdateSetting(key, value)
	m.sendUpdateMember()
}

func (m *ComponentMaster) UpdateSettings(setting map[string]string) {
	for key, value := range setting {
		m.thisMember.UpdateSetting(key, value)
	}

	m.sendUpdateMember()
}

func (m *ComponentMaster) Stop() {
	if m.isClient() {
		m.sendRemove(m.thisMember.NodeID)
	}

	if m.cancel != nil {
		m.cancel()
	}

	clog.Debugf("[Stop] NodeID = %s is unregister", m.App().NodeID())
}

func (m *ComponentMaster) isMaster() bool {
	return m.App().NodeID() == m.masterID
}

func (m *ComponentMaster) isClient() bool {
	return m.App().NodeID() != m.masterID
}

func (m *ComponentMaster) buildSubject(subject string) string {
	return fmt.Sprintf(subject, m.prefix, m.masterID)
}

func (m *ComponentMaster) init() {
	m.loadThisMember()

	// Build subjects
	m.registerSubject = m.buildSubject("cherry.%s.discovery.%s.register")
	m.addSubject = m.buildSubject("cherry.%s.discovery.%s.add")
	m.updateSubject = m.buildSubject("cherry.%s.discovery.%s.update")
	m.removeSubject = m.buildSubject("cherry.%s.discovery.%s.remove")
	m.heartbeatSubject = m.buildSubject("cherry.%s.discovery.%s.heartbeat")

	// Build cancel context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Node init
	m.masterInit()
	m.clientInit()

	clog.Infof("[init] Discovery = %s is running.", m.Mode())
}

func (m *ComponentMaster) loadThisMember() {
	// Get nats config
	config := cprofile.GetConfig("cluster").GetConfig(m.Mode())
	if config.LastError() != nil {
		clog.Fatalf("[loadMember] Nats config not found. err = %v", config.LastError())
	}

	// Set prefix
	m.prefix = config.GetString("prefix", "node")

	// Get master node id
	m.masterID = config.GetString("master_node_id")
	if m.masterID == "" {
		clog.Fatal("[loadMember] Master node id not in config.")
	}

	// Build this member
	m.thisMember = NewMemberWithApp(m.App())

	// add member
	m.AddMember(m.thisMember)
}

func (m *ComponentMaster) masterInit() {
	if m.isMaster() {
		// subscribe register message
		m.registerSubscribe()
		// subscribe update message
		m.updateSubscribe()
		// subscribe remove message
		m.removeSubscribe()

		// subscribe heartbeat message
		go m.heartbeatCheck()
		m.heartbeatSubscribe()
	}
}

func (m *ComponentMaster) clientInit() {
	if m.isClient() {
		// receive registered node
		m.addSubscribe()
		// subscribe update message
		m.updateSubscribe()
		// subscribe remove message
		m.removeSubscribe()

		// send register&heartbeat message to master node
		go m.clientTicker()
	}
}

func (m *ComponentMaster) addSubscribe() {
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

func (m *ComponentMaster) clientTicker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			clog.Info("[clientTicker] Is exit.")
			return
		case <-ticker.C:
			if !m.App().Running() {
				clog.Info("[clientTicker] Waiting for the application to change its running state.")
				continue
			}

			// When the Application is in the running state, a registration message is sent
			if m.sendHeartbeat2Master() {
				m.sendRegister2Master()
			}
		}
	}
}

func (m *ComponentMaster) sendHeartbeat2Master() bool {
	nodeIDBytes, err := m.NodeID2Bytes(m.thisMember.NodeID)
	if err != nil {
		clog.Warnf("[sendHeartbeat2Master] NodeID to bytes error. err = %s", err)
		return false
	}

	reqID := cnats.NewStringReqID()
	rspData, err := cnats.RequestSync(reqID, m.heartbeatSubject, nodeIDBytes)
	if err != nil {
		clog.Warnf("[sendHeartbeat2Master] Fail. master = %s, err = %s", m.masterID, err)
		return false
	}

	return len(rspData) > 0
}

func (m *ComponentMaster) sendRegister2Master() {
	memberBytes, err := m.member2Bytes(m.thisMember)
	if err != nil {
		clog.Warnf("[sendRegister2Master] member marshal error. err = %s", err)
		return
	}

	reqID := cnats.NewStringReqID()
	rspData, err := cnats.RequestSync(reqID, m.registerSubject, memberBytes)
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
		if member.NodeID == m.thisMember.NodeID {
			continue
		}

		if _, ok := m.GetMember(member.GetNodeID()); !ok {
			m.AddMember(member)
		}
	}
}

func (m *ComponentMaster) sendUpdateMember() {
	memberBytes, err := m.member2Bytes(m.thisMember)
	if err != nil {
		clog.Warnf("[UpdateMember] member marshal error. err = %s", err)
		return
	}

	err = cnats.Publish(m.updateSubject, memberBytes)
	if err != nil {
		clog.Warnf("[UpdateMember] Fail. master = %s, err = %s", m.masterID, err)
		return
	}
}

func (m *ComponentMaster) sendRemove(nodeID string) {
	nodeBytes, err := m.NodeID2Bytes(nodeID)
	if err != nil {
		clog.Warnf("[sendRemove] NodeID2Bytes error. err = %s", err)
		return
	}

	err = cnats.Publish(m.removeSubject, nodeBytes)
	if err != nil {
		clog.Warnf("[sendRemove] Publish fail. err = %s", err)
	}
}

func (m *ComponentMaster) heartbeatCheck() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			clog.Info(" heartbeat check is exit.")
			return
		case <-ticker.C:
			m.checkMemberTimeout()
		}
	}
}

func (m *ComponentMaster) checkMemberTimeout() {
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
			m.sendRemove(protoMember.NodeID)
		}

		return true
	})
}

func (m *ComponentMaster) registerSubscribe() {
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

func (m *ComponentMaster) heartbeatSubscribe() {
	m.subscribe(m.heartbeatSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[heartbeatSubscribe] bytes to NodeID error. err = %v", err)
			return
		}

		// Get reply ReqID
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

func (m *ComponentMaster) replyMemberList(msg *nats.Msg) {
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

func (m *ComponentMaster) sendAdd(member *cproto.Member) {
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

func (m *ComponentMaster) removeSubscribe() {
	m.subscribe(m.removeSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[removeSubscribe] bytes to NodeID error. err = %s", err)
			return
		}

		if m.isMaster() && nodeID == m.App().NodeID() {
			return
		}

		// remove member
		m.ComponentDefault.RemoveMember(nodeID)
	})
}

func (m *ComponentMaster) updateSubscribe() {
	m.subscribe(m.updateSubject, func(msg *nats.Msg) {
		member, err := m.bytes2Member(msg.Data)
		if err != nil {
			clog.Warnf("[updateSubscribe] bytes to member error. err = %s", err)
			return
		}

		if member.NodeID == m.App().NodeID() {
			return
		}

		// update member
		m.ComponentDefault.UpdateMember(member)
	})

}

func (m *ComponentMaster) member2Bytes(member cfacade.IMember) ([]byte, error) {
	return m.App().Serializer().Marshal(member)
}

func (m *ComponentMaster) memberList2Bytes() ([]byte, error) {
	memberList := &cproto.MemberList{}
	m.memberMap.Range(func(key, value any) bool {
		protoMember := value.(*cproto.Member)
		memberList.List = append(memberList.List, protoMember)
		return true
	})

	return m.App().Serializer().Marshal(memberList)
}

func (m *ComponentMaster) bytes2MemberList(data []byte) (*cproto.MemberList, error) {
	memberList := &cproto.MemberList{}
	err := m.App().Serializer().Unmarshal(data, memberList)
	if err != nil {
		return nil, err
	}

	return memberList, nil
}

func (m *ComponentMaster) NodeID2Bytes(nodeID string) ([]byte, error) {
	nodeProto := &cproto.NodeID{
		Value: nodeID,
	}

	return m.App().Serializer().Marshal(nodeProto)
}

func (m *ComponentMaster) bytes2Member(data []byte) (*cproto.Member, error) {
	member := &cproto.Member{}
	err := m.App().Serializer().Unmarshal(data, member)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (m *ComponentMaster) bytes2NodeID(data []byte) (string, error) {
	nodeIDProto := &cproto.NodeID{}
	err := m.App().Serializer().Unmarshal(data, nodeIDProto)
	if err != nil {
		return "", err
	}

	return nodeIDProto.Value, nil
}

func (m *ComponentMaster) subscribe(subject string, cb nats.MsgHandler) {
	err := cnats.Subscribe(subject, cb)
	if err != nil {
		clog.Warnf("[subscribe] fail. subject = %s, err = %s", subject, err)
		return
	}
}

func NewMemberWithApp(app cherryFacade.IApplication) *cproto.Member {
	// Default timeout is 3 seconds
	clusterHeartbeatTimeout := app.Settings().GetInt64("cluster_heartbeat_timeout", 3) * ctime.MillisecondsPerSecond

	return &cproto.Member{
		NodeID:           app.NodeID(),
		NodeType:         app.NodeType(),
		Address:          app.RpcAddress(),
		LastAt:           ctime.Now().ToMillisecond(),
		HeartbeatTimeout: clusterHeartbeatTimeout,
		Settings:         make(map[string]string),
	}
}
