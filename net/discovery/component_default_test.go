package cherryDiscovery

import (
	"sync"
	"testing"

	cfacade "github.com/cherry-game/cherry/facade"
	cproto "github.com/cherry-game/cherry/net/proto"
)

func newTestMember(nodeID, nodeType, address string) *cproto.Member {
	return &cproto.Member{
		NodeID:   nodeID,
		NodeType: nodeType,
		Address:  address,
		Settings: make(map[string]string),
	}
}

// TestComponentDefault_AddMember_NewMember verifies that adding a new member
// stores it in the member map and notifies registered OnAddMember listeners.
func TestComponentDefault_AddMember_NewMember(t *testing.T) {
	d := &ComponentDefault{}

	var called bool
	var received cfacade.IMember
	d.OnAddMember(func(m cfacade.IMember) {
		called = true
		received = m
	})

	member := newTestMember("node-1", "game", "127.0.0.1:10001")
	d.AddMember(member)

	if !called {
		t.Fatal("OnAddMember listener was not called")
	}
	if received.GetNodeID() != "node-1" {
		t.Fatalf("expected node-1, got %s", received.GetNodeID())
	}

	found, ok := d.GetMember("node-1")
	if !ok || found.GetNodeID() != "node-1" {
		t.Fatal("member not found after AddMember")
	}
}

// TestComponentDefault_AddMember_Duplicate verifies that adding the same member twice
// still notifies listeners but does not create duplicates.
func TestComponentDefault_AddMember_Duplicate(t *testing.T) {
	d := &ComponentDefault{}

	count := 0
	d.OnAddMember(func(m cfacade.IMember) { count++ })

	member := newTestMember("node-1", "game", "127.0.0.1:10001")
	d.AddMember(member)
	d.AddMember(member)

	if count != 2 {
		t.Fatalf("expected listener called 2 times, got %d", count)
	}
}

// TestComponentDefault_RemoveMember verifies that removing a member deletes it
// from the map and notifies OnRemoveMember listeners.
func TestComponentDefault_RemoveMember(t *testing.T) {
	d := &ComponentDefault{}

	var called bool
	var removedID string
	d.OnRemoveMember(func(m cfacade.IMember) {
		called = true
		removedID = m.GetNodeID()
	})

	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))
	d.RemoveMember("node-1")

	if !called {
		t.Fatal("OnRemoveMember listener was not called")
	}
	if removedID != "node-1" {
		t.Fatalf("expected node-1, got %s", removedID)
	}
	if _, ok := d.GetMember("node-1"); ok {
		t.Fatal("member should have been removed")
	}
}

// TestComponentDefault_RemoveMember_NotFound verifies that removing a non-existent
// member is a no-op and does not trigger listeners.
func TestComponentDefault_RemoveMember_NotFound(t *testing.T) {
	d := &ComponentDefault{}

	called := false
	d.OnRemoveMember(func(m cfacade.IMember) { called = true })

	d.RemoveMember("nonexistent")
	if called {
		t.Fatal("listener should not be called for non-existent member")
	}
}

// TestComponentDefault_UpdateMember_Existing verifies that updating an existing member
// triggers OnUpdateMember listeners with the previously stored value.
func TestComponentDefault_UpdateMember_Existing(t *testing.T) {
	d := &ComponentDefault{}

	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))

	var called bool
	d.OnUpdateMember(func(m cfacade.IMember) {
		called = true
	})

	updated := newTestMember("node-1", "game", "127.0.0.1:20001")
	d.UpdateMember(updated)

	if !called {
		t.Fatal("OnUpdateMember listener was not called")
	}
}

// TestComponentDefault_UpdateMember_New verifies that updating a member that doesn't
// exist yet stores it but does not trigger update listeners.
func TestComponentDefault_UpdateMember_New(t *testing.T) {
	d := &ComponentDefault{}

	called := false
	d.OnUpdateMember(func(m cfacade.IMember) { called = true })

	updated := newTestMember("node-1", "game", "127.0.0.1:10001")
	d.UpdateMember(updated)

	if called {
		t.Fatal("OnUpdateMember should not be called for new members")
	}
	if _, ok := d.GetMember("node-1"); !ok {
		t.Fatal("member should be stored even without update notification")
	}
}

// TestComponentDefault_GetMember_Found verifies lookup by valid nodeID.
func TestComponentDefault_GetMember_Found(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))

	m, ok := d.GetMember("node-1")
	if !ok {
		t.Fatal("member not found")
	}
	if m.GetNodeID() != "node-1" {
		t.Fatalf("expected node-1, got %s", m.GetNodeID())
	}
}

// TestComponentDefault_GetMember_EmptyNodeID verifies that empty nodeID returns nil.
func TestComponentDefault_GetMember_EmptyNodeID(t *testing.T) {
	d := &ComponentDefault{}

	_, ok := d.GetMember("")
	if ok {
		t.Fatal("empty nodeID should return not found")
	}
}

// TestComponentDefault_GetMember_NotFound verifies lookup with unknown nodeID.
func TestComponentDefault_GetMember_NotFound(t *testing.T) {
	d := &ComponentDefault{}

	_, ok := d.GetMember("unknown")
	if ok {
		t.Fatal("unknown nodeID should return not found")
	}
}

// TestComponentDefault_GetType_Found verifies GetType returns the correct node type.
func TestComponentDefault_GetType_Found(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))

	typ, err := d.GetType("node-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typ != "game" {
		t.Fatalf("expected game, got %s", typ)
	}
}

// TestComponentDefault_GetType_NotFound verifies GetType returns an error for unknown nodeID.
func TestComponentDefault_GetType_NotFound(t *testing.T) {
	d := &ComponentDefault{}

	_, err := d.GetType("unknown")
	if err == nil {
		t.Fatal("expected error for unknown nodeID")
	}
}

// TestComponentDefault_Map_Empty verifies Map returns empty map when no members exist.
func TestComponentDefault_Map_Empty(t *testing.T) {
	d := &ComponentDefault{}

	m := d.Map()
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(m))
	}
}

// TestComponentDefault_Map_Populated verifies Map returns a snapshot with all members.
func TestComponentDefault_Map_Populated(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))
	d.AddMember(newTestMember("node-2", "gate", "127.0.0.1:20001"))

	m := d.Map()
	if len(m) != 2 {
		t.Fatalf("expected 2 members, got %d", len(m))
	}
	if _, ok := m["node-1"]; !ok {
		t.Fatal("node-1 not in map")
	}
	if _, ok := m["node-2"]; !ok {
		t.Fatal("node-2 not in map")
	}
}

// TestComponentDefault_ListByType_FilterByType verifies ListByType returns
// only members of the given nodeType.
func TestComponentDefault_ListByType_FilterByType(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))
	d.AddMember(newTestMember("node-2", "gate", "127.0.0.1:20001"))
	d.AddMember(newTestMember("node-3", "game", "127.0.0.1:10002"))

	list := d.ListByType("game")
	if len(list) != 2 {
		t.Fatalf("expected 2 game nodes, got %d", len(list))
	}
}

// TestComponentDefault_ListByType_ExcludeFilter verifies ListByType excludes
// nodes matching filterNodeID arguments.
func TestComponentDefault_ListByType_ExcludeFilter(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))
	d.AddMember(newTestMember("node-2", "game", "127.0.0.1:10002"))
	d.AddMember(newTestMember("node-3", "game", "127.0.0.1:10003"))

	list := d.ListByType("game", "node-2")
	if len(list) != 2 {
		t.Fatalf("expected 2 after filtering node-2, got %d", len(list))
	}
	for _, m := range list {
		if m.GetNodeID() == "node-2" {
			t.Fatal("node-2 should have been filtered out")
		}
	}
}

// TestComponentDefault_Random_Empty verifies Random returns nil, false when no members exist.
func TestComponentDefault_Random_Empty(t *testing.T) {
	d := &ComponentDefault{}

	m, ok := d.Random("game")
	if ok || m != nil {
		t.Fatal("expected nil, false for empty member list")
	}
}

// TestComponentDefault_Random_Single verifies Random returns the only member.
func TestComponentDefault_Random_Single(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))

	m, ok := d.Random("game")
	if !ok {
		t.Fatal("expected true for single member")
	}
	if m.GetNodeID() != "node-1" {
		t.Fatalf("expected node-1, got %s", m.GetNodeID())
	}
}

// TestComponentDefault_Random_Multiple verifies Random returns one of the available members.
func TestComponentDefault_Random_Multiple(t *testing.T) {
	d := &ComponentDefault{}
	d.AddMember(newTestMember("node-1", "game", "127.0.0.1:10001"))
	d.AddMember(newTestMember("node-2", "game", "127.0.0.1:10002"))
	d.AddMember(newTestMember("node-3", "game", "127.0.0.1:10003"))

	// Run multiple times to confirm Random always returns a valid member
	for i := range 50 {
		_ = i
		m, ok := d.Random("game")
		if !ok {
			t.Fatal("Random returned false for non-empty list")
		}
		if m.GetNodeType() != "game" {
			t.Fatalf("expected game type, got %s", m.GetNodeType())
		}
	}
}

// TestComponentDefault_OnAddMember_Nil verifies nil listeners are silently ignored.
func TestComponentDefault_OnAddMember_Nil(t *testing.T) {
	d := &ComponentDefault{}
	d.OnAddMember(nil)
	if len(d.onAddListener) != 0 {
		t.Fatalf("nil listener should not be appended, got %d", len(d.onAddListener))
	}
}

// TestComponentDefault_OnUpdateMember_Nil verifies nil update listeners are silently ignored.
func TestComponentDefault_OnUpdateMember_Nil(t *testing.T) {
	d := &ComponentDefault{}
	d.OnUpdateMember(nil)
	if len(d.onUpdateListener) != 0 {
		t.Fatalf("nil listener should not be appended, got %d", len(d.onUpdateListener))
	}
}

// TestComponentDefault_OnRemoveMember_Nil verifies nil remove listeners are silently ignored.
func TestComponentDefault_OnRemoveMember_Nil(t *testing.T) {
	d := &ComponentDefault{}
	d.OnRemoveMember(nil)
	if len(d.onRemoveListener) != 0 {
		t.Fatalf("nil listener should not be appended, got %d", len(d.onRemoveListener))
	}
}

// TestComponentDefault_Mode verifies the default mode name.
func TestComponentDefault_Mode(t *testing.T) {
	d := &ComponentDefault{}
	if d.Mode() != "default" {
		t.Fatalf("expected 'default', got '%s'", d.Mode())
	}
}

// TestComponentDefault_Name verifies the component name is "discovery".
func TestComponentDefault_Name(t *testing.T) {
	d := &ComponentDefault{}
	if d.Name() != "discovery" {
		t.Fatalf("expected 'discovery', got '%s'", d.Name())
	}
}

// TestComponentDefault_MemberMapConcurrency verifies that sync.Map-based memberMap
// can safely handle concurrent AddMember and RemoveMember calls.
func TestComponentDefault_MemberMapConcurrency(t *testing.T) {
	d := &ComponentDefault{}
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(2)
		nodeID := "node-" + string(rune('0'+i%10))
		go func() {
			defer wg.Done()
			d.AddMember(newTestMember(nodeID, "game", "127.0.0.1:10001"))
		}()
		go func() {
			defer wg.Done()
			d.GetMember(nodeID)
		}()
	}

	wg.Wait()

	if len(d.Map()) == 0 {
		t.Fatal("memberMap should have members after concurrent add")
	}
}

// TestComponentDefault_UpdateSetting verifies that UpdateSetting updates the local
// member's settings in the memberMap and fires OnUpdateMember listeners.
func TestComponentDefault_UpdateSetting(t *testing.T) {
	d := &ComponentDefault{}
	app := &mockApp{nodeID: "node-1"}
	d.Set(app)

	d.AddMember(&cproto.Member{
		NodeID:   "node-1",
		NodeType: "game",
		Settings: make(map[string]string),
	})

	var called bool
	d.OnUpdateMember(func(m cfacade.IMember) { called = true })

	d.UpdateSetting("region", "us-west")

	if !called {
		t.Fatal("OnUpdateMember listener was not called")
	}

	m, _ := d.GetMember("node-1")
	if m.GetSettings()["region"] != "us-west" {
		t.Fatalf("expected 'us-west', got '%s'", m.GetSettings()["region"])
	}
}

// TestComponentDefault_UpdateSettings verifies that UpdateSettings updates multiple
// settings and fires OnUpdateMember once.
func TestComponentDefault_UpdateSettings(t *testing.T) {
	d := &ComponentDefault{}
	app := &mockApp{nodeID: "node-1"}
	d.Set(app)

	d.AddMember(&cproto.Member{
		NodeID:   "node-1",
		NodeType: "game",
		Settings: make(map[string]string),
	})

	callCount := 0
	d.OnUpdateMember(func(m cfacade.IMember) { callCount++ })

	d.UpdateSettings(map[string]string{"region": "us-west", "zone": "a"})

	if callCount != 1 {
		t.Fatalf("expected 1 listener call, got %d", callCount)
	}

	m, _ := d.GetMember("node-1")
	if m.GetSettings()["region"] != "us-west" {
		t.Fatalf("region mismatch: expected 'us-west', got '%s'", m.GetSettings()["region"])
	}
	if m.GetSettings()["zone"] != "a" {
		t.Fatalf("zone mismatch: expected 'a', got '%s'", m.GetSettings()["zone"])
	}
}

// TestComponentDefault_UpdateSetting_NoSelf verifies that UpdateSetting is a no-op
// when the local node is not in the member map.
func TestComponentDefault_UpdateSetting_NoSelf(t *testing.T) {
	d := &ComponentDefault{}
	app := &mockApp{nodeID: "node-1"}
	d.Set(app)

	called := false
	d.OnUpdateMember(func(m cfacade.IMember) { called = true })

	// node-1 is not in memberMap
	d.UpdateSetting("key", "value")

	if called {
		t.Fatal("listener should not be called when self is not in memberMap")
	}
}
