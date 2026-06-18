package cherryDiscovery

import (
	"testing"

	cfacade "github.com/cherry-game/cherry/facade"
)

// testDiscoveryComponent is a minimal IDiscoveryComponent for testing Register/GetDiscovery.
type testDiscoveryComponent struct {
	mode string
}

func (t *testDiscoveryComponent) Mode() string                             { return t.mode }
func (t *testDiscoveryComponent) Name() string                             { return "test" }
func (t *testDiscoveryComponent) App() cfacade.IApplication                { return nil }
func (t *testDiscoveryComponent) Set(cfacade.IApplication)                 {}
func (t *testDiscoveryComponent) Init()                                    {}
func (t *testDiscoveryComponent) OnAfterInit()                             {}
func (t *testDiscoveryComponent) OnBeforeStop()                            {}
func (t *testDiscoveryComponent) OnStop()                                  {}
func (t *testDiscoveryComponent) Map() map[string]cfacade.IMember          { return nil }
func (t *testDiscoveryComponent) ListByType(string, ...string) []cfacade.IMember { return nil }
func (t *testDiscoveryComponent) Random(string) (cfacade.IMember, bool)    { return nil, false }
func (t *testDiscoveryComponent) GetMember(string) (cfacade.IMember, bool) { return nil, false }
func (t *testDiscoveryComponent) UpdateSetting(string, string)              {}
func (t *testDiscoveryComponent) UpdateSettings(map[string]string)          {}
func (t *testDiscoveryComponent) OnAddMember(cfacade.MemberListener)       {}
func (t *testDiscoveryComponent) OnUpdateMember(cfacade.MemberListener)    {}
func (t *testDiscoveryComponent) OnRemoveMember(cfacade.MemberListener)    {}

// TestGetDiscovery_Found verifies that a registered component can be looked up by mode.
func TestGetDiscovery_Found(t *testing.T) {
	comp := &testDiscoveryComponent{mode: "test-mode"}
	Register(comp)

	result, err := GetDiscovery("test-mode")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != comp {
		t.Fatal("returned component does not match registered one")
	}
}

// TestGetDiscovery_NotFound verifies that looking up an unregistered mode returns an error.
func TestGetDiscovery_NotFound(t *testing.T) {
	_, err := GetDiscovery("nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistered mode")
	}
}

// TestGetDiscovery_Builtins verifies that the built-in "default" and "nats" modes
// are registered by init() and discoverable.
func TestGetDiscovery_Builtins(t *testing.T) {
	comp, err := GetDiscovery("default")
	if err != nil {
		t.Fatalf("default mode not registered: %v", err)
	}
	if comp == nil {
		t.Fatal("default component is nil")
	}

	comp, err = GetDiscovery("nats")
	if err != nil {
		t.Fatalf("nats mode not registered: %v", err)
	}
	if comp == nil {
		t.Fatal("nats component is nil")
	}
}
