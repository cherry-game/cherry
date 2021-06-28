package cherryConnector

import (
	"encoding/json"
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/agent"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/packet"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
	"time"
)

type (
	// Component (连接器组件适用于前端节点)
	Component struct {
		facade.Component
		cherryAgent.Options
		ConnectStat      *ConnectStat
		sessionComponent *cherrySession.SessionComponent
		handlerComponent *cherryHandler.Component
		connector        facade.IConnector
	}
)

var (
	DefaultHeartbeat = 60 * time.Second
)

func NewTCPComponent(address string) *Component {
	opt := cherryAgent.Options{
		Heartbeat:        DefaultHeartbeat,
		DataCompression:  false,
		PacketDecoder:    cherryPacket.NewPomeloDecoder(),
		PacketEncoder:    cherryPacket.NewPomeloEncoder(),
		Serializer:       cherrySerializer.NewProtobuf(),
		PacketListener:   make(map[cherryPacket.Type]cherryAgent.PacketListener),
		RPCHandler:       nil,
		OnCreateListener: make([]cherryAgent.SessionListener, 0),
		OnCloseListener:  make([]cherryAgent.SessionListener, 0),
	}

	ws := NewTCP(address)

	return NewComponentWithOpt(opt, ws)
}

func NewWSComponent(address string) *Component {
	opt := cherryAgent.Options{
		Heartbeat:        DefaultHeartbeat,
		DataCompression:  false,
		PacketDecoder:    cherryPacket.NewPomeloDecoder(),
		PacketEncoder:    cherryPacket.NewPomeloEncoder(),
		Serializer:       cherrySerializer.NewProtobuf(),
		PacketListener:   make(map[cherryPacket.Type]cherryAgent.PacketListener),
		RPCHandler:       nil,
		OnCreateListener: make([]cherryAgent.SessionListener, 0),
		OnCloseListener:  make([]cherryAgent.SessionListener, 0),
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
	p.sessionComponent = p.App().Find(cherryConst.SessionComponent).(*cherrySession.SessionComponent)
	if p.sessionComponent == nil {
		panic("please preload session component.")
	}

	p.handlerComponent = p.App().Find(cherryConst.HandlerComponent).(*cherryHandler.Component)
	if p.handlerComponent == nil {
		panic("preload handler component please.")
	}

	p.OnCreateSession(func(s *cherrySession.Session) (next bool, err error) {
		// increase connect stat
		p.ConnectStat.IncreaseConn()

		s.Debugf("session on create. %s", s.String())

		return true, nil
	})

	p.OnCloseSession(func(s *cherrySession.Session) (next bool, err error) {
		// decrease connect stat
		p.ConnectStat.DecreaseConn()
		return true, nil
	})

	//add packet listener
	p.AddPacketHandle(cherryPacket.Handshake, p.handshake)
	p.AddPacketHandle(cherryPacket.HandshakeAck, p.handshakeACK)
	p.AddPacketHandle(cherryPacket.Heartbeat, p.heartbeat)
	p.AddPacketHandle(cherryPacket.Data, p.handData)

	// default on connect
	p.connector.OnConnect(func(conn facade.INetConn) {
		// create new session
		session := p.sessionComponent.Create(cherrySession.NextSID(), p.App().NodeId())

		// create agent
		agent := cherryAgent.NewAgent(p.Options, session, conn)

		session.SetNetwork(agent)

		// run agent
		agent.Run()
	})

	// new goroutine
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

func (p *Component) OnCreateSession(listener ...cherryAgent.SessionListener) {
	p.Options.OnCreateListener = append(p.Options.OnCreateListener, listener...)
}

func (p *Component) OnCloseSession(listener ...cherryAgent.SessionListener) {
	p.Options.OnCloseListener = append(p.Options.OnCloseListener, listener...)
}

func (p *Component) AddPacketHandle(typ cherryPacket.Type, listener cherryAgent.PacketListener) {
	p.Options.PacketListener[typ] = listener
}

func (p *Component) handshake(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	data := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat": p.Heartbeat.Seconds(),
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	bytes, err := agent.PacketEncoder.Encode(cherryPacket.Handshake, jsonData)
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

func (p *Component) handshakeACK(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	agent.SetStatus(cherryAgent.Working)

	agent.Session.Debug("request handshakeACK.")
}

func (p *Component) heartbeat(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	bytes, err := agent.PacketEncoder.Encode(cherryPacket.Heartbeat, nil)
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

func (p *Component) handData(agent *cherryAgent.Agent, pkg *cherryPacket.Packet) {
	if agent.Status() != cherryAgent.Working {
		agent.Session.Warnf("status is not working. status[%d]", agent.Status())
		return
	}

	msg, err := cherryMessage.Decode(pkg.Data)
	if err != nil {
		agent.Session.Warnf("packet decode error. data[%s], error[%s].", pkg.Data, err)
		return
	}

	route, err := cherryRoute.Decode(msg.Route)
	if err != nil {
		agent.Session.Warnf("route decode error. route[%s], error[%s]", msg.Route, err)
		return
	}

	p.handlerComponent.PostMessage(agent, route, msg)
}
