package cherryDiscovery

import (
	"context"
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

// registerRequired is a marker byte sent back to unknown heartbeat senders
// to trigger a full registration request.
var registerRequired = []byte{0x01}

// ComponentMaster implements NATS-based master/worker discovery.
//
// Protocol overview:
//  1. Start a single master node first.
//  2. Worker nodes send a register message (cherry.<prefix>.discovery.<masterID>.register) to the master.
//  3. The master replies with the full member list and broadcasts the new member via add subject.
//  4. Workers periodically send heartbeat messages; the master removes members that time out.
//  5. Workers broadcast remove messages on shutdown.
//
// natsSubjects holds the NATS subject strings built from prefix and masterID.
// thisMember is the local node's member info, synced with the master on setting changes.
// ctx/cancel control the lifecycle of background goroutines (ticker, heartbeat check).
type (
	ComponentMaster struct {
		ComponentDefault
		natsSubjects
		thisMember *cproto.Member     // local node's member info, updated on setting changes
		masterID   string             // the designated master node ID from config
		ctx        context.Context    // lifecycle context for background goroutines
		cancel     context.CancelFunc // cancel func
	}

	natsSubjects struct {
		prefix           string // NATS subject prefix (default "node")
		registerSubject  string // master: receive register; worker: send register
		addSubject       string // master: broadcast add; worker: receive add
		updateSubject    string // both: receive update notifications
		removeSubject    string // both: receive remove notifications
		heartbeatSubject string // master: receive heartbeat; worker: send heartbeat
	}
)

func NewMaster() ComponentMaster {
	return ComponentMaster{}
}

func (m *ComponentMaster) Mode() string {
	return "nats"
}

// Init sets up the master component: loads local member info, builds NATS subjects,
// and starts the appropriate subscriptions and background loops based on role.
func (m *ComponentMaster) Init() {
	m.init()
}

// UpdateSetting updates a single setting on this member and broadcasts the change.
func (m *ComponentMaster) UpdateSetting(key, value string) {
	m.thisMember.UpdateSetting(key, value)
	m.sendUpdateMember()
}

// UpdateSettings updates multiple settings on this member and broadcasts the change.
func (m *ComponentMaster) UpdateSettings(setting map[string]string) {
	for key, value := range setting {
		m.thisMember.UpdateSetting(key, value)
	}

	m.sendUpdateMember()
}

// Stop performs a graceful shutdown: workers send a remove message,
// then the lifecycle context is cancelled to stop all background loops.
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

// buildSubject formats a NATS subject template with prefix and masterID.
func (m *ComponentMaster) buildSubject(subject string) string {
	return fmt.Sprintf(subject, m.prefix, m.masterID)
}

// init performs the full initialization sequence:
// load config → build subjects → init based on role (master or worker).
func (m *ComponentMaster) init() {
	m.loadThisMember()

	// Build NATS subjects for all discovery messages
	m.registerSubject = m.buildSubject("cherry.%s.discovery.%s.register")
	m.addSubject = m.buildSubject("cherry.%s.discovery.%s.add")
	m.updateSubject = m.buildSubject("cherry.%s.discovery.%s.update")
	m.removeSubject = m.buildSubject("cherry.%s.discovery.%s.remove")
	m.heartbeatSubject = m.buildSubject("cherry.%s.discovery.%s.heartbeat")

	// Build cancel context for background goroutine lifecycle
	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.masterInit()
	m.clientInit()

	clog.Infof("[init] Discovery = %s is running. [isMaster = %s, nodeID = %s]", m.Mode(), m.isMaster(), m.App().NodeID())
}

// loadThisMember reads NATS/member config from profile and constructs the local member.
// Config path: cluster.<mode>  e.g. cluster.nats
func (m *ComponentMaster) loadThisMember() {
	config := cprofile.GetConfig("cluster").GetConfig(m.Mode())
	if config.LastError() != nil {
		clog.Fatalf("[loadMember] Nats config not found. err = %v", config.LastError())
	}

	m.prefix = config.GetString("prefix", "node")

	m.masterID = config.GetString("master_node_id")
	if m.masterID == "" {
		clog.Fatal("[loadMember] Master node id not in config.")
	}

	m.thisMember = NewMemberWithApp(m.App())

	// Register self in the local member table
	m.AddMember(m.thisMember)
}

// masterInit sets up subscriptions that only the master node needs:
// register, update, remove, and heartbeat (with timeout checking).
func (m *ComponentMaster) masterInit() {
	if !m.isMaster() {
		return
	}

	m.registerSubscribe()
	m.updateSubscribe()
	m.removeSubscribe()

	// Master runs heartbeat timeout detection in background
	go m.heartbeatCheck()
	m.heartbeatSubscribe()
}

// clientInit sets up subscriptions that worker nodes need:
// add (to learn about new members), update, remove, and periodic heartbeat/register.
func (m *ComponentMaster) clientInit() {
	if !m.isClient() {
		return
	}

	m.addSubscribe()
	m.updateSubscribe()
	m.removeSubscribe()

	// Workers periodically heartbeat and re-register
	go m.clientTicker()
}

// addSubscribe handles "member added" broadcasts from the master.
// New members are added to the local table unless they are self.
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

// clientTicker runs on worker nodes: every second it sends a heartbeat.
// If the heartbeat reply indicates the master doesn't know this node (non-empty reply),
// a full registration is triggered. Waits for the application to reach running state.
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

			// Heartbeat first; if master replies with registerRequired marker,
			// send a full registration to sync the member list.
			if m.sendHeartbeat2Master() {
				m.sendRegister2Master()
			}
		}
	}
}

// sendHeartbeat2Master sends a heartbeat request to the master.
// Returns true if the master replied (meaning this node is known),
// or true if the reply contains registerRequired (unknown node, re-register needed).
// Returns false on error (master unreachable).
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

// sendRegister2Master sends a full registration to the master and processes
// the member list reply, adding any members not yet known locally.
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

// sendUpdateMember broadcasts updated member info (typically settings changes)
// to all nodes via the update subject.
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

// sendRemove publishes a remove notification to the master/peers for the given nodeID.
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

// heartbeatCheck runs on the master node: every second it iterates all members
// and removes those whose LastAt + HeartbeatTimeout exceeds the current time.
// Also sends a remove broadcast for each timed-out member.
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

// checkMemberTimeout scans all known members and evicts those whose
// last heartbeat timestamp exceeds their configured heartbeat timeout.
// The local node (thisMember) is never evicted.
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

// registerSubscribe handles incoming register requests on the master.
// Steps: 1) stamp heartbeat time on the new member, 2) add to local table,
// 3) broadcast add to all workers, 4) reply with the full member list.
func (m *ComponentMaster) registerSubscribe() {
	m.subscribe(m.registerSubject, func(msg *nats.Msg) {
		newMember, err := m.bytes2Member(msg.Data)
		if err != nil {
			clog.Warnf("[registerSubscribe] bytes to Member error. err = %s", err)
			return
		}

		// Initialize heartbeat timestamp to now so the new member isn't
		// immediately considered timed out.
		newMember.LastAt = ctime.Now().ToMillisecond()
		m.AddMember(newMember)
		m.sendAdd(newMember)
		m.replyMemberList(msg)
	})
}

// heartbeatSubscribe handles heartbeat pings on the master.
// Known members get their LastAt timestamp updated and an empty reply.
// Unknown members get a registerRequired marker reply to trigger full registration.
func (m *ComponentMaster) heartbeatSubscribe() {
	m.subscribe(m.heartbeatSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[heartbeatSubscribe] bytes to NodeID error. err = %v", err)
			return
		}

		reqID := msg.Header.Get(cnats.REQ_ID)

		if value, found := m.GetMember(nodeID); found {
			// Known node: refresh heartbeat timestamp
			if protoMember, ok := value.(*cproto.Member); ok {
				protoMember.LastAt = ctime.Now().ToMillisecond()
			}
			cnats.ReplySync(reqID, msg.Reply, nil)
		} else {
			// Unknown node: signal it to send a full registration
			cnats.ReplySync(reqID, msg.Reply, registerRequired)
		}
	})
}

// replyMemberList sends the full member list back to a requesting node
// as a reply to the given NATS message.
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

// sendAdd broadcasts a member add notification to all nodes.
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

// removeSubscribe handles incoming remove notifications.
// Both master and workers ignore self-remove to prevent accidental self-deletion.
func (m *ComponentMaster) removeSubscribe() {
	m.subscribe(m.removeSubject, func(msg *nats.Msg) {
		nodeID, err := m.bytes2NodeID(msg.Data)
		if err != nil {
			clog.Warnf("[removeSubscribe] bytes to NodeID error. err = %s", err)
			return
		}

		// Never remove self via this path — applies to both master and workers
		if nodeID == m.App().NodeID() {
			return
		}

		m.ComponentDefault.RemoveMember(nodeID)
	})
}

// updateSubscribe handles incoming member update notifications.
// Ignores updates from self to avoid echo loops.
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

		m.ComponentDefault.UpdateMember(member)
	})

}

// --- Serialization helpers ---

func (m *ComponentMaster) member2Bytes(member cfacade.IMember) ([]byte, error) {
	return m.App().Serializer().Marshal(member)
}

func (m *ComponentMaster) bytes2Member(data []byte) (*cproto.Member, error) {
	member := &cproto.Member{}
	err := m.App().Serializer().Unmarshal(data, member)
	if err != nil {
		return nil, err
	}
	return member, nil
}

// memberList2Bytes builds a MemberList proto from the current memberMap and serializes it.
func (m *ComponentMaster) memberList2Bytes() ([]byte, error) {
	memberList := &cproto.MemberList{}
	m.memberMap.Range(func(key, value any) bool {
		memberList.List = append(memberList.List, value.(*cproto.Member))
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
	return m.App().Serializer().Marshal(&cproto.NodeID{Value: nodeID})
}

func (m *ComponentMaster) bytes2NodeID(data []byte) (string, error) {
	nodeIDProto := &cproto.NodeID{}
	err := m.App().Serializer().Unmarshal(data, nodeIDProto)
	if err != nil {
		return "", err
	}
	return nodeIDProto.Value, nil
}

// subscribe is a thin wrapper around cnats.Subscribe that logs failures as warnings.
func (m *ComponentMaster) subscribe(subject string, cb nats.MsgHandler) {
	err := cnats.Subscribe(subject, cb)
	if err != nil {
		clog.Warnf("[subscribe] fail. subject = %s, err = %s", subject, err)
		return
	}
}

// NewMemberWithApp creates a *cproto.Member populated from the application's node identity.
// HeartbeatTimeout is read from the "cluster_heartbeat_timeout" setting (default 3s).
func NewMemberWithApp(app cfacade.IApplication) *cproto.Member {
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
