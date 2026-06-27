package simple

import (
	"encoding/binary"
	"net"
	"time"

	cnet "github.com/cherry-game/cherry/extend/net"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nuid"
	"go.uber.org/zap/zapcore"
)

type (
	actor struct {
		cactor.Base
		agentActorID   string
		connectors     []cfacade.IConnector
		onNewAgentFunc OnNewAgentFunc
	}

	OnNewAgentFunc func(newAgent *Agent)
)

// NewActor creates a new simple parser Actor with the given agent actor id.
func NewActor(agentActorID string) *actor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &actor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *actor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
}

// Load starts the parser: creates the agent actor and starts all registered connectors.
func (p *actor) Load(app cfacade.IApplication) {
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
func (p *actor) AddConnector(connector cfacade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

// Connectors returns the list of registered connectors.
func (p *actor) Connectors() []cfacade.IConnector {
	return p.connectors
}

// AddNodeRoute maps a mid (message id) to a NodeRoute for routing incoming messages.
func (p *actor) AddNodeRoute(mid uint32, nodeRoute *NodeRoute) {
	AddNodeRoute(mid, nodeRoute)
}

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent actor
func (p *actor) defaultOnConnectFunc(conn net.Conn) {
	session := &cproto.Session{
		Sid:       nuid.Next(),
		AgentPath: p.Path().String(),
		Data:      map[string]string{},
		Ip:        cnet.GetIPV4(conn.RemoteAddr()),
	}

	agent := NewAgent(p.App(), conn, session)

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(agent)
	}

	BindSID(agent)
	agent.Run()
}

// SetOnNewAgent sets the callback that is invoked when a new agent connection is established.
func (p *actor) SetOnNewAgent(fn OnNewAgentFunc) {
	p.onNewAgentFunc = fn
}

// SetHeartbeatTime sets the heartbeat interval for agent connections.
func (p *actor) SetHeartbeatTime(t time.Duration) {
	SetHeartbeatTime(t)
}

// SetWriteBacklog sets the size of the write and pending channel buffers.
func (p *actor) SetWriteBacklog(backlog int) {
	SetWriteBacklog(backlog)
}

// SetEndian sets the byte order used for encoding/decoding message headers.
func (p *actor) SetEndian(e binary.ByteOrder) {
	SetEndian(e)
}

// SetOnDataRoute sets the callback that handles routing of incoming data messages.
func (*actor) SetOnDataRoute(fn DataRouteFunc) {
	if fn != nil {
		onDataRouteFunc = fn
	}
}

func (p *actor) response(rsp *cproto.PomeloResponse) {
	agent, found := GetAgentWithSID(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[response] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	agent.Response(rsp.Mid, rsp.Data)
}
