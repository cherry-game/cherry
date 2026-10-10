package cherryProto

import "testing"

// TestSessionWriteWithNilData covers a session that crossed the cluster with
// nothing to carry: proto3 does not serialize an empty map, so the receiver
// ends up with a nil Data map. Remote handlers do write session data, so a
// write must allocate on demand instead of panicking with
// "assignment to entry in nil map".
func TestSessionWriteWithNilData(t *testing.T) {
	session := &Session{Sid: "sid-nil"}

	session.Add("player_id", 100)
	if got := session.GetString("player_id"); got != "100" {
		t.Fatalf("Add on a nil data map: got %q", got)
	}

	session.Set("game_node_id", "game-1")
	if got := session.GetString("game_node_id"); got != "game-1" {
		t.Fatalf("Set on a nil data map: got %q", got)
	}

	session.Remove("game_node_id")
	if session.Contains("game_node_id") {
		t.Fatal("Remove on a nil data map failed")
	}
}

// TestSessionMidIsIndependentFromData pins down that mid is a plain proto
// field: it must never end up in the data map, so a session without business
// data stays empty and mid still round-trips on its own.
func TestSessionMidIsIndependentFromData(t *testing.T) {
	session := &Session{Sid: "sid-mid"}

	session.SetMID(1234)

	if got := session.GetMID(); got != 1234 {
		t.Fatalf("GetMID = %d, want 1234", got)
	}

	if len(session.Data) != 0 {
		t.Fatalf("mid must not be written into Data: %v", session.Data)
	}
}
