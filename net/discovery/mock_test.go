package cherryDiscovery

import (
	"google.golang.org/protobuf/proto"

	cfacade "github.com/cherry-game/cherry/facade"
)

// mockSerializer implements cfacade.ISerializer using protobuf marshal/unmarshal.
type mockSerializer struct{}

func (mockSerializer) Marshal(v any) ([]byte, error)  { return proto.Marshal(v.(proto.Message)) }
func (mockSerializer) Unmarshal(b []byte, v any) error { return proto.Unmarshal(b, v.(proto.Message)) }
func (mockSerializer) Name() string                            { return "protobuf" }

// mockApp is a minimal cfacade.IApplication implementation for testing.
// Only the fields needed by each test need to be set; unused interface methods
// return zero values.
type mockApp struct {
	cfacade.INode
	nodeID    string
	nodeType  string
	rpcAddr   string
	settings  cfacade.ProfileJSON
	isRunning bool
	serial    cfacade.ISerializer
}

func (a *mockApp) NodeID() string                    { return a.nodeID }
func (a *mockApp) NodeType() string                  { return a.nodeType }
func (a *mockApp) RpcAddress() string                { return a.rpcAddr }
func (a *mockApp) Settings() cfacade.ProfileJSON     { return a.settings }
func (a *mockApp) Running() bool                     { return a.isRunning }
func (a *mockApp) Serializer() cfacade.ISerializer   { return a.serial }
func (a *mockApp) DieChan() chan bool                { return nil }
func (a *mockApp) IsFrontend() bool                  { return false }
func (a *mockApp) Register(...cfacade.IComponent)    {}
func (a *mockApp) Find(string) cfacade.IComponent    { return nil }
func (a *mockApp) Remove(string) cfacade.IComponent  { return nil }
func (a *mockApp) All() []cfacade.IComponent         { return nil }
func (a *mockApp) OnShutdown(...func())              {}
func (a *mockApp) Startup()                          {}
func (a *mockApp) Shutdown()                         {}
func (a *mockApp) Discovery() cfacade.IDiscovery     { return nil }
func (a *mockApp) Cluster() cfacade.ICluster         { return nil }
func (a *mockApp) ActorSystem() cfacade.IActorSystem { return nil }
func (a *mockApp) Address() string                   { return "" }
func (a *mockApp) Enabled() bool                     { return true }
