package simple

import (
	"encoding/binary"
	"net"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nuid"
	"go.uber.org/zap/zapcore"
)

type (
	Actor struct {
		cactor.Base
		agentActorID   string
		connectors     []cfacade.IConnector
		onNewAgentFunc OnNewAgentFunc
	}

	OnNewAgentFunc func(newAgent *Agent)
)

// NewActor creates a new simple parser Actor with the given agent actor id.
func NewActor(agentActorID string) *Actor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &Actor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *Actor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
}

// Load starts the parser: creates the agent actor and starts all registered connectors.
func (p *Actor) Load(app cfacade.IApplication) {
	if len(p.connectors) < 1 {
		panic("Connectors is nil. Please call the AddConnector(...) method add IConnector.")
	}

	//  Create agent actor
	if _, err := app.ActorSystem().CreateActor(p.agentActorID, p); err != nil {
		clog.Panicf("Create agent actor fail. err = %+v", err)
	}

	for _, connector := range p.connectors {
		connector.OnConnect(p.defaultOnConnectFunc)
		go connector.Start() // start connector!
	}
}

// AddConnector registers a connector that will be started when the parser loads.
func (p *Actor) AddConnector(connector cfacade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

// Connectors returns the list of registered connectors.
func (p *Actor) Connectors() []cfacade.IConnector {
	return p.connectors
}

// AddNodeRoute maps a mid (message id) to a NodeRoute for routing incoming messages.
func (p *Actor) AddNodeRoute(mid uint32, nodeRoute *NodeRoute) {
	AddNodeRoute(mid, nodeRoute)
}

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent actor
func (p *Actor) defaultOnConnectFunc(conn net.Conn) {
	agent := NewAgent(p.App(), conn, nuid.Next(), p.Path().String())

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(agent)
	}

	BindSID(agent)
	agent.Run()
}

// SetOnNewAgent sets the callback that is invoked when a new agent connection is established.
func (p *Actor) SetOnNewAgent(fn OnNewAgentFunc) {
	p.onNewAgentFunc = fn
}

// SetHeartbeatTime sets the heartbeat interval for agent connections.
func (p *Actor) SetHeartbeatTime(t time.Duration) {
	SetHeartbeatTime(t)
}

// SetWriteBacklog sets the size of the write and pending channel buffers.
func (p *Actor) SetWriteBacklog(backlog int) {
	SetWriteBacklog(backlog)
}

// SetEndian sets the byte order used for encoding/decoding message headers.
func (p *Actor) SetEndian(e binary.ByteOrder) {
	SetEndian(e)
}

// SetOnDataRoute sets the callback that handles routing of incoming data messages.
func (*Actor) SetOnDataRoute(fn DataRouteFunc) {
	if fn != nil {
		onDataRouteFunc = fn
	}
}

func (p *Actor) response(rsp *cproto.PomeloResponse) {
	agent, found := GetAgentWithSID(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[response] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	agent.Response(rsp.Mid, rsp.Data)
}
