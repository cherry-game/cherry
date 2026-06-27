package pomelo

import (
	"net"
	"time"

	ccode "github.com/cherry-game/cherry/code"
	cnet "github.com/cherry-game/cherry/extend/net"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	pomeloMessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	ppacket "github.com/cherry-game/cherry/net/parser/pomelo/packet"
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
		onInitFunc     func()
	}

// OnNewAgentFunc is called when a new agent connection is established.
	OnNewAgentFunc func(newAgent *Agent)
)

// NewActor creates a new pomelo parser Actor with the given agent actor id.
func NewActor(agentActorID string) *Actor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &Actor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
		onInitFunc:   nil,
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *Actor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
	p.Remote().Register(PushFuncName, p.push)
	p.Remote().Register(KickFuncName, p.kick)
	p.Remote().Register(BroadcastName, p.broadcast)

	if p.onInitFunc != nil {
		p.onInitFunc()
	}
}

// SetOnInitFunc sets a callback that is invoked during Actor initialization.
func (p *Actor) SetOnInitFunc(fn func()) {
	p.onInitFunc = fn
}

// Load starts the parser: initializes the command, creates the agent actor,
// and starts all registered connectors.
func (p *Actor) Load(app cfacade.IApplication) {
	if len(p.connectors) < 1 {
		panic("connectors is nil. Please call the AddConnector(...) method add IConnector.")
	}

	cmd.init(app)

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

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent actor
func (p *Actor) defaultOnConnectFunc(conn net.Conn) {
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

// SetDictionary sets the route-to-code dictionary used for route compression.
func (*Actor) SetDictionary(dict map[string]uint16) {
	pomeloMessage.SetDictionary(dict)
}

// SetDataCompression enables or disables data compression for message encoding.
func (*Actor) SetDataCompression(compression bool) {
	pomeloMessage.SetDataCompression(compression)
}

// SetWriteBacklog sets the size of the write and pending channel buffers.
func (*Actor) SetWriteBacklog(size int) {
	cmd.writeBacklog = size
}

// SetHeartbeat sets the heartbeat interval. Values less than 1 second default to 60 seconds.
func (*Actor) SetHeartbeat(t time.Duration) {
	if t.Seconds() < 1 {
		t = 60 * time.Second
	}
	cmd.heartbeatTime = t
}

// SetSysData sets a system data key-value pair that is sent during handshake.
func (*Actor) SetSysData(key string, value interface{}) {
	cmd.sysData[key] = value
}

// SetOnNewAgent sets the callback that is invoked when a new agent connection is established.
func (p *Actor) SetOnNewAgent(fn OnNewAgentFunc) {
	p.onNewAgentFunc = fn
}

// SetOnDataRoute sets the callback that handles routing of incoming data messages.
// If fn is nil the call is silently ignored and the default handler is kept.
func (*Actor) SetOnDataRoute(fn DataRouteFunc) {
	if fn != nil {
		cmd.onDataRouteFunc = fn
	}
}

// SetOnPacket registers a handler for the given packet type. Only installed once per type;
// subsequent calls for the same type are ignored.
func (*Actor) SetOnPacket(typ ppacket.Type, fn PacketFunc) {
	cmd.onPacketFuncMap[typ] = fn
}

func (p *Actor) response(rsp *cproto.PomeloResponse) {
	agent, found := GetAgentWithSID(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[response] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	if ccode.IsOK(rsp.Code) {
		agent.ResponseMID(rsp.Mid, rsp.Data, false)
	} else {
		errRsp := &cproto.Response{
			Code: rsp.Code,
		}
		agent.ResponseMID(rsp.Mid, errRsp, true)
	}
}

func (p *Actor) push(rsp *cproto.PomeloPush) {
	if rsp.Sid != "" || rsp.Uid > 0 {
		if agent, found := GetAgent(rsp.Sid, rsp.Uid); found {
			agent.Push(rsp.Route, rsp.Data)
		}

		return
	}
}

func (p *Actor) kick(rsp *cproto.PomeloKick) {
	if rsp.Sid != "" || rsp.Uid > 0 {
		if agent, found := GetAgent(rsp.Sid, rsp.Uid); found {
			agent.Kick(rsp.Reason, rsp.Close)
		}

		return
	}
}

func (p *Actor) broadcast(rsp *cproto.PomeloBroadcast) {
	switch rsp.PushType {
	case cproto.PomeloBroadcast_AllUID:
		{
			ForeachAgent(func(agent *Agent) {
				if agent.IsBind() {
					agent.Push(rsp.Route, rsp.Data)
				}
			})

			return
		}
	case cproto.PomeloBroadcast_UID:
		{
			for _, uid := range rsp.UidList {
				if agent, found := GetAgentWithUID(uid); found {
					agent.Push(rsp.Route, rsp.Data)
				}
			}

			return
		}
	}
}
