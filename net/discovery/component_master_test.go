package cherryDiscovery

import (
	"testing"

	ctime "github.com/cherry-game/cherry/extend/time"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// newTestMaster creates a ComponentMaster with a mock app pre-configured
// for testing. The mock provides a protobuf serializer and basic node identity.
func newTestMaster(nodeID, masterID string) *ComponentMaster {
	m := &ComponentMaster{
		masterID: masterID,
	}
	app := &mockApp{
		nodeID:   nodeID,
		nodeType: "game",
		rpcAddr:  "127.0.0.1:10001",
		serial:   mockSerializer{},
	}
	m.Set(app)
	return m
}

// TestComponentMaster_Mode verifies the nats mode name.
func TestComponentMaster_Mode(t *testing.T) {
	m := &ComponentMaster{}
	if m.Mode() != "nats" {
		t.Fatalf("expected 'nats', got '%s'", m.Mode())
	}
}

// TestComponentMaster_IsMaster verifies isMaster returns true when the app's
// NodeID matches the configured masterID.
func TestComponentMaster_IsMaster(t *testing.T) {
	m := newTestMaster("master-1", "master-1")
	if !m.isMaster() {
		t.Fatal("expected isMaster=true when NodeID == masterID")
	}
}

// TestComponentMaster_IsMaster_False verifies isMaster returns false when NodeID
// differs from masterID.
func TestComponentMaster_IsMaster_False(t *testing.T) {
	m := newTestMaster("worker-1", "master-1")
	if m.isMaster() {
		t.Fatal("expected isMaster=false when NodeID != masterID")
	}
}

// TestComponentMaster_IsClient verifies isClient returns true for worker nodes.
func TestComponentMaster_IsClient(t *testing.T) {
	m := newTestMaster("worker-1", "master-1")
	if !m.isClient() {
		t.Fatal("expected isClient=true for worker node")
	}
}

// TestComponentMaster_IsClient_False verifies isClient returns false for the master node.
func TestComponentMaster_IsClient_False(t *testing.T) {
	m := newTestMaster("master-1", "master-1")
	if m.isClient() {
		t.Fatal("expected isClient=false for master node")
	}
}

// TestComponentMaster_BuildSubject verifies buildSubject formats the NATS subject
// template with the configured prefix and masterID.
func TestComponentMaster_BuildSubject(t *testing.T) {
	m := &ComponentMaster{}
	m.prefix = "node"
	m.masterID = "master-1"

	result := m.buildSubject("cherry.%s.discovery.%s.register")
	expected := "cherry.node.discovery.master-1.register"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

// TestComponentMaster_Member2Bytes_Bytes2Member_Roundtrip verifies that serializing
// a member and deserializing it back preserves all fields.
func TestComponentMaster_Member2Bytes_Bytes2Member_Roundtrip(t *testing.T) {
	m := newTestMaster("node-1", "master-1")

	original := &cproto.Member{
		NodeID:   "node-1",
		NodeType: "game",
		Address:  "127.0.0.1:10001",
		Settings: map[string]string{"key": "value"},
	}

	bytes, err := m.member2Bytes(original)
	if err != nil {
		t.Fatalf("member2Bytes failed: %v", err)
	}

	restored, err := m.bytes2Member(bytes)
	if err != nil {
		t.Fatalf("bytes2Member failed: %v", err)
	}

	if restored.NodeID != original.NodeID {
		t.Fatalf("NodeID mismatch: expected %s, got %s", original.NodeID, restored.NodeID)
	}
	if restored.NodeType != original.NodeType {
		t.Fatalf("NodeType mismatch: expected %s, got %s", original.NodeType, restored.NodeType)
	}
	if restored.Address != original.Address {
		t.Fatalf("Address mismatch: expected %s, got %s", original.Address, restored.Address)
	}
	if restored.Settings["key"] != "value" {
		t.Fatalf("Settings mismatch: expected 'value', got '%s'", restored.Settings["key"])
	}
}

// TestComponentMaster_Bytes2Member_InvalidData verifies that deserializing garbage
// data returns an error.
func TestComponentMaster_Bytes2Member_InvalidData(t *testing.T) {
	m := newTestMaster("node-1", "master-1")

	_, err := m.bytes2Member([]byte("not valid protobuf"))
	if err == nil {
		t.Fatal("expected error for invalid protobuf data")
	}
}

// TestComponentMaster_NodeID2Bytes_Bytes2NodeID_Roundtrip verifies roundtrip
// serialization of a nodeID.
func TestComponentMaster_NodeID2Bytes_Bytes2NodeID_Roundtrip(t *testing.T) {
	m := newTestMaster("node-1", "master-1")

	bytes, err := m.NodeID2Bytes("test-node")
	if err != nil {
		t.Fatalf("NodeID2Bytes failed: %v", err)
	}

	nodeID, err := m.bytes2NodeID(bytes)
	if err != nil {
		t.Fatalf("bytes2NodeID failed: %v", err)
	}
	if nodeID != "test-node" {
		t.Fatalf("expected 'test-node', got '%s'", nodeID)
	}
}

// TestComponentMaster_Bytes2NodeID_InvalidData verifies that deserializing garbage
// returns an error.
func TestComponentMaster_Bytes2NodeID_InvalidData(t *testing.T) {
	m := newTestMaster("node-1", "master-1")

	_, err := m.bytes2NodeID([]byte("not valid protobuf"))
	if err == nil {
		t.Fatal("expected error for invalid protobuf data")
	}
}

// TestComponentMaster_MemberList2Bytes_Bytes2MemberList_Roundtrip verifies
// serialization roundtrip of a full member list.
func TestComponentMaster_MemberList2Bytes_Bytes2MemberList_Roundtrip(t *testing.T) {
	m := newTestMaster("node-1", "master-1")

	// Populate memberMap with test members
	m.memberMap.Store("node-1", &cproto.Member{NodeID: "node-1", NodeType: "game"})
	m.memberMap.Store("node-2", &cproto.Member{NodeID: "node-2", NodeType: "gate"})

	bytes, err := m.memberList2Bytes()
	if err != nil {
		t.Fatalf("memberList2Bytes failed: %v", err)
	}

	memberList, err := m.bytes2MemberList(bytes)
	if err != nil {
		t.Fatalf("bytes2MemberList failed: %v", err)
	}

	if len(memberList.List) != 2 {
		t.Fatalf("expected 2 members, got %d", len(memberList.List))
	}
}

// TestComponentMaster_Bytes2MemberList_InvalidData verifies deserialization error for
// invalid member list data.
func TestComponentMaster_Bytes2MemberList_InvalidData(t *testing.T) {
	m := newTestMaster("node-1", "master-1")

	_, err := m.bytes2MemberList([]byte("invalid"))
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}

// TestComponentMaster_CheckMemberTimeout_TimeoutRemoved verifies that members
// whose heartbeat has expired are removed from the member map.
func TestComponentMaster_CheckMemberTimeout_TimeoutRemoved(t *testing.T) {
	m := newTestMaster("master-1", "master-1")
	m.thisMember = &cproto.Member{NodeID: "master-1"}

	now := ctime.Now().ToMillisecond()

	// Member with expired heartbeat: LastAt is far in the past
	m.memberMap.Store("worker-1", &cproto.Member{
		NodeID:           "worker-1",
		LastAt:           now - 10000, // 10 seconds ago
		HeartbeatTimeout: 3000,        // 3 second timeout
	})

	m.checkMemberTimeout()

	if _, ok := m.GetMember("worker-1"); ok {
		t.Fatal("timeout member should have been removed")
	}
}

// TestComponentMaster_CheckMemberTimeout_ActiveKept verifies that members with
// recent heartbeats are NOT removed.
func TestComponentMaster_CheckMemberTimeout_ActiveKept(t *testing.T) {
	m := newTestMaster("master-1", "master-1")
	m.thisMember = &cproto.Member{NodeID: "master-1"}

	now := ctime.Now().ToMillisecond()

	// Member with recent heartbeat
	m.memberMap.Store("worker-1", &cproto.Member{
		NodeID:           "worker-1",
		LastAt:           now,   // just now
		HeartbeatTimeout: 3000,  // 3 second timeout
	})

	m.checkMemberTimeout()

	if _, ok := m.GetMember("worker-1"); !ok {
		t.Fatal("active member should not have been removed")
	}
}

// TestComponentMaster_CheckMemberTimeout_SelfSkipped verifies that the local node
// (thisMember) is never removed during timeout checking, even if its own heartbeat
// timestamp is expired.
func TestComponentMaster_CheckMemberTimeout_SelfSkipped(t *testing.T) {
	m := newTestMaster("master-1", "master-1")
	m.thisMember = &cproto.Member{
		NodeID:           "master-1",
		LastAt:           0,    // ancient — would normally be timeout
		HeartbeatTimeout: 3000,
	}
	m.memberMap.Store("master-1", m.thisMember)

	m.checkMemberTimeout()

	if _, ok := m.GetMember("master-1"); !ok {
		t.Fatal("self should never be removed by timeout check")
	}
}

// TestComponentMaster_CheckMemberTimeout_TypeError verifies that a non-Member value
// in the member map is skipped gracefully (logged, not panicked).
func TestComponentMaster_CheckMemberTimeout_TypeError(t *testing.T) {
	m := newTestMaster("master-1", "master-1")
	m.thisMember = &cproto.Member{NodeID: "master-1"}

	// Store a non-Member value (this shouldn't happen in practice, but we handle it)
	m.memberMap.Store("bad-node", "not a member")

	// Should not panic
	m.checkMemberTimeout()
}

// TestComponentMaster_UpdateSetting verifies that UpdateSetting updates the
// local thisMember Settings map.
func TestComponentMaster_UpdateSetting(t *testing.T) {
	m := newTestMaster("node-1", "master-1")
	m.thisMember = &cproto.Member{
		NodeID:   "node-1",
		Settings: make(map[string]string),
	}

	m.UpdateSetting("region", "us-west")

	if m.thisMember.Settings["region"] != "us-west" {
		t.Fatalf("expected 'us-west', got '%s'", m.thisMember.Settings["region"])
	}
}

// TestComponentMaster_UpdateSettings verifies that UpdateSettings updates
// multiple settings at once.
func TestComponentMaster_UpdateSettings(t *testing.T) {
	m := newTestMaster("node-1", "master-1")
	m.thisMember = &cproto.Member{
		NodeID:   "node-1",
		Settings: make(map[string]string),
	}

	m.UpdateSettings(map[string]string{
		"region": "us-west",
		"zone":   "a",
	})

	if m.thisMember.Settings["region"] != "us-west" {
		t.Fatalf("region mismatch: expected 'us-west', got '%s'", m.thisMember.Settings["region"])
	}
	if m.thisMember.Settings["zone"] != "a" {
		t.Fatalf("zone mismatch: expected 'a', got '%s'", m.thisMember.Settings["zone"])
	}
}
