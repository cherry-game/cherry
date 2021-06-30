package cherryConnector

import (
	"encoding/json"
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/agent"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	"github.com/cherry-game/cherry/net/session"
	"time"
)

type (
	// Component (连接器组件适用于前端节点)
	Component struct {
		facade.Component
		cherryAgent.Options
		ConnectStat      *ConnectStat
		sessionComponent *cherrySession.Component
		handlerComponent *cherryHandler.Component
		connector        facade.IConnector
	}
)

var (
	DefaultHeartbeat = 60 * time.Second
)

func NewTCPComponent(address string) *Component {
	opt := cherryAgent.Options{
		Heartbeat:       DefaultHeartbeat,
		DataCompression: false,
		PacketListener:  make(map[cherryPacket.Type]cherryAgent.PacketListener),
		RPCHandler:      nil,
	}

	ws := NewTCP(address)

	return NewComponentWithOpt(opt, ws)
}

func NewWSComponent(address string) *Component {
	opt := cherryAgent.Options{
		Heartbeat:       DefaultHeartbeat,
		DataCompression: false,
		PacketListener:  make(map[cherryPacket.Type]cherryAgent.PacketListener),
		RPCHandler:      nil,
	}

	ws := NewWS(address)

	return NewComponentWithOpt(opt, ws)
}

func NewComponentWithOpt(opts cherryAgent.Options, connector facade.IConnector) *Component {
	return &Component{
		Options:     opts,
		connector:   connector,
		ConnectStat: &ConnectStat{},
	}
}

func (p *Component) Name() string {
	return cherryConst.ConnectorPomeloComponent
}

func (p *Component) Init() {

}

func (p *Component) OnAfterInit() {
	p.sessionComponent = p.App().Find(cherryConst.SessionComponent).(*cherrySession.Component)
	if p.sessionComponent == nil {
		panic("session component must be preloaded.")
	}

	p.handlerComponent = p.App().Find(cherryConst.HandlerComponent).(*cherryHandler.Component)
	if p.handlerComponent == nil {
		panic("handler component must be preloaded.")
	}

	// set rpc handler
	p.Options.RPCHandler = p.handlerComponent.PostMessage

	p.sessionComponent.AddOnCreate(func(s *cherrySession.Session) (next bool) {
		p.ConnectStat.IncreaseConn()
		s.Debugf("session on create. %s", s.String())
		p.ConnectStat.PrintInfo()
		return true
	})

	p.sessionComponent.AddOnClose(func(s *cherrySession.Session) (next bool) {
		p.ConnectStat.DecreaseConn()
		p.ConnectStat.PrintInfo()
		return true
	})

	//add packet listener
	p.AddPacketHandle(cherryPacket.Handshake, p.handshake)
	p.AddPacketHandle(cherryPacket.HandshakeAck, p.handshakeACK)
	p.AddPacketHandle(cherryPacket.Heartbeat, p.heartbeat)
	p.AddPacketHandle(cherryPacket.Data, p.handData)

	// default on connect
	p.connector.OnConnect(func(conn facade.INetConn) {
		// create agent
		agent := cherryAgent.NewAgent(p.App(), p.Options, conn)

		// create new session
		session := p.sessionComponent.Create(agent)
		agent.Session = session

		// run agent
		agent.Run()
	})

	// new goroutine for connector
	go p.connector.OnStart()
}

func (p *Component) OnStop() {
	if p.sessionComponent != nil {
		p.sessionComponent.CloseAll()
	}

	if p.connector != nil {
		p.connector.OnStop()
	}
}

func (p *Component) AddPacketHandle(typ cherryPacket.Type, listener cherryAgent.PacketListener) {
	p.Options.PacketListener[typ] = listener
}

func (p *Component) handshake(agent *cherryAgent.Agent, _ facade.IPacket) {
	data := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat": agent.Options.Heartbeat.Seconds(),
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	bytes, err := p.App().PacketEncode(byte(cherryPacket.Handshake), jsonData)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	agent.SetStatus(cherryAgent.WaitAck)
	err = agent.SendRaw(bytes)
	if err != nil {
		cherryLogger.Error(err)
	}

	agent.Session.Debugf("request handshake. data[%v]", data)
}

func (p *Component) handshakeACK(agent *cherryAgent.Agent, _ facade.IPacket) {
	agent.SetStatus(cherryAgent.Working)

	agent.Session.Debug("request handshakeACK.")
}

func (p *Component) heartbeat(agent *cherryAgent.Agent, _ facade.IPacket) {
	bytes, err := p.App().PacketEncode(byte(cherryPacket.Heartbeat), nil)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	err = agent.SendRaw(bytes)
	if err != nil {
		cherryLogger.Error(err)
	}
	agent.Session.Debug("request heartbeat.")
}

func (p *Component) handData(agent *cherryAgent.Agent, pkg facade.IPacket) {
	if agent.Status() != cherryAgent.Working {
		agent.Session.Warnf("status is not working. status[%d]", agent.Status())
		return
	}

	msg, err := cherryMessage.Decode(pkg.Data())
	if err != nil {
		agent.Session.Warnf("packet decode error. data[%s], error[%s].", pkg.Data, err)
		return
	}

	p.handlerComponent.PostMessage(agent.Session, msg)
}
