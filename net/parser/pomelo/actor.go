package pomelo

import (
	"net"
	"time"

	ccode "github.com/cherry-game/cherry/code"
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
	actor struct {
		cactor.Base
		agentActorID   string
		connectors     []cfacade.IConnector
		onNewAgentFunc OnNewAgentFunc
		onInitFunc     func()
	}

	OnNewAgentFunc func(newAgent *Agent)
)

func NewActor(agentActorID string) *actor {
	if agentActorID == "" {
		panic("agentActorID is empty.")
	}

	parser := &actor{
		agentActorID: agentActorID,
		connectors:   make([]cfacade.IConnector, 0),
		onInitFunc:   nil,
	}

	return parser
}

// OnInit Actor初始化前触发该函数
func (p *actor) OnInit() {
	p.Remote().Register(ResponseFuncName, p.response)
	p.Remote().Register(PushFuncName, p.push)
	p.Remote().Register(KickFuncName, p.kick)
	p.Remote().Register(BroadcastName, p.broadcast)

	if p.onInitFunc != nil {
		p.onInitFunc()
	}
}

func (p *actor) SetOnInitFunc(fn func()) {
	p.onInitFunc = fn
}

func (p *actor) Load(app cfacade.IApplication) {
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

func (p *actor) AddConnector(connector cfacade.IConnector) {
	p.connectors = append(p.connectors, connector)
}

func (p *actor) Connectors() []cfacade.IConnector {
	return p.connectors
}

// defaultOnConnectFunc 创建新连接时，通过当前agentActor创建child agent actor
func (p *actor) defaultOnConnectFunc(conn net.Conn) {
	session := &cproto.Session{
		Sid:       nuid.Next(),
		AgentPath: p.Path().String(),
		Data:      map[string]string{},
	}

	agent := NewAgent(p.App(), conn, session)

	if p.onNewAgentFunc != nil {
		p.onNewAgentFunc(&agent)
	}

	BindSID(&agent)
	agent.Run()
}

func (*actor) SetDictionary(dict map[string]uint16) {
	pomeloMessage.SetDictionary(dict)
}

func (*actor) SetDataCompression(compression bool) {
	pomeloMessage.SetDataCompression(compression)
}

func (*actor) SetWriteBacklog(size int) {
	cmd.writeBacklog = size
}

func (*actor) SetHeartbeat(t time.Duration) {
	if t.Seconds() < 1 {
		t = 60 * time.Second
	}
	cmd.heartbeatTime = t
}

func (*actor) SetSysData(key string, value interface{}) {
	cmd.sysData[key] = value
}

func (p *actor) SetOnNewAgent(fn OnNewAgentFunc) {
	p.onNewAgentFunc = fn
}

func (*actor) SetOnDataRoute(fn DataRouteFunc) {
	if fn != nil {
		cmd.onDataRouteFunc = fn
	}
}

func (*actor) SetOnPacket(typ ppacket.Type, fn PacketFunc) {
	cmd.onPacketFuncMap[typ] = fn
}

func (p *actor) response(rsp *cproto.PomeloResponse) {
	agent, found := GetAgent(rsp.Sid)
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

func (p *actor) push(rsp *cproto.PomeloPush) {
	agent, found := GetAgent(rsp.Sid)
	if !found {
		if clog.PrintLevel(zapcore.DebugLevel) {
			clog.Debugf("[push] Not found agent. [rsp = %+v]", rsp)
		}
		return
	}

	agent.Push(rsp.Route, rsp.Data)
}

func (p *actor) kick(rsp *cproto.PomeloKick) {
	agent, found := GetAgentWithUID(rsp.Uid)
	if !found {
		agent, found = GetAgent(rsp.Sid)
	}

	if found {
		agent.Kick(rsp.Reason, rsp.Close)
	}
}

func (p *actor) broadcast(rsp *cproto.PomeloBroadcastPush) {
	if rsp.AllUID {
		ForeachAgent(func(agent *Agent) {
			if agent.IsBind() {
				agent.Push(rsp.Route, rsp.Data)
			}
		})
	} else {
		for _, uid := range rsp.UidList {
			if agent, found := GetAgentWithUID(uid); found {
				agent.Push(rsp.Route, rsp.Data)
			}
		}
	}
}
