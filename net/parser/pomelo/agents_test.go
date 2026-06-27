package pomelo

import (
	"testing"

	cfacade "github.com/cherry-game/cherry/facade"
	cproto "github.com/cherry-game/cherry/net/proto"
)

func newTestAgent(sid string) *Agent {
	return &Agent{
		session: &cproto.Session{
			Sid: cfacade.SID(sid),
		},
	}
}

func TestBindSIDAndGetAgentWithSID(t *testing.T) {
	agent := newTestAgent("sid-1")
	BindSID(agent)

	got, found := GetAgentWithSID("sid-1")
	if !found {
		t.Fatal("GetAgentWithSID not found")
	}
	if got.SID() != "sid-1" {
		t.Fatalf("expected sid-1, got %s", got.SID())
	}
}

func TestBindAndGetAgentWithUID(t *testing.T) {
	agent := newTestAgent("sid-2")
	BindSID(agent)

	_, err := Bind("sid-2", 100)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	got, found := GetAgentWithUID(100)
	if !found {
		t.Fatal("GetAgentWithUID not found")
	}
	if got.UID() != 100 {
		t.Fatalf("expected uid 100, got %d", got.UID())
	}
}

func TestBindDuplicateLogin(t *testing.T) {
	old := newTestAgent("old-sid")
	BindSID(old)
	Bind("old-sid", 200)

	neu := newTestAgent("new-sid")
	BindSID(neu)

	oldAgent, err := Bind("new-sid", 200)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	if oldAgent == nil {
		t.Fatal("expected old agent for duplicate login")
	}
	if oldAgent.SID() != "old-sid" {
		t.Fatalf("expected old-sid, got %s", oldAgent.SID())
	}

	got, _ := GetAgentWithUID(200)
	if got.SID() != "new-sid" {
		t.Fatalf("expected new-sid to own uid, got %s", got.SID())
	}
}

func TestUnbind(t *testing.T) {
	agent := newTestAgent("sid-3")
	BindSID(agent)
	Bind("sid-3", 300)

	Unbind("sid-3")

	_, found := GetAgentWithSID("sid-3")
	if found {
		t.Fatal("agent should be removed after Unbind")
	}
	_, found = GetAgentWithUID(300)
	if found {
		t.Fatal("uid mapping should be removed after Unbind")
	}
}

func TestUnbindKeepsNewSIDMapping(t *testing.T) {
	a1 := newTestAgent("s1")
	BindSID(a1)
	Bind("s1", 400)

	a2 := newTestAgent("s2")
	BindSID(a2)
	Bind("s2", 400)

	Unbind("s1")

	got, found := GetAgentWithUID(400)
	if !found {
		t.Fatal("uid should still be bound to s2")
	}
	if got.SID() != "s2" {
		t.Fatalf("expected s2, got %s", got.SID())
	}
}

func TestGetAgent(t *testing.T) {
	agent := newTestAgent("sid-5")
	BindSID(agent)
	Bind("sid-5", 500)

	got, found := GetAgent("sid-5", 0)
	if !found || got.SID() != "sid-5" {
		t.Fatal("GetAgent by sid failed")
	}

	got, found = GetAgent("", 500)
	if !found || got.UID() != 500 {
		t.Fatal("GetAgent by uid failed")
	}

	_, found = GetAgent("", 0)
	if found {
		t.Fatal("GetAgent with empty args should not find")
	}
}

func TestCountAndForeach(t *testing.T) {
	before := Count()

	for i := 0; i < 3; i++ {
		sid := cfacade.SID("sid-foreach-" + string(rune('a'+i)))
		BindSID(newTestAgent(string(sid)))
	}

	if Count() < before+3 {
		t.Fatalf("Count should be at least %d, got %d", before+3, Count())
	}

	seen := 0
	ForeachAgent(func(a *Agent) {
		seen++
	})
	if seen < before+3 {
		t.Fatalf("ForeachAgent saw %d, expected at least %d", seen, before+3)
	}
}
