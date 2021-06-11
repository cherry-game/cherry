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
)

type (
	// ConnectorComponent (连接器组件适用于前端节点)
	ConnectorComponent struct {
		facade.Component
		cherryAgent.AgentOpt
		ConnectStat      *ConnectStat
		sessionComponent *cherrySession.SessionComponent
		handlerComponent *cherryHandler.HandlerComponent
		connector        facade.IConnector
		packetHandles    map[byte]cherryAgent.PacketListener
	}
)

func NewTCPComponent(address string) *ConnectorComponent {
	opt := cherryAgent.AgentOpt{
		Heartbeat:        60,
		DataCompression:  false,
		PacketDecoder:    cherryPacket.NewPomeloDecoder(),
		PacketEncoder:    cherryPacket.NewPomeloEncoder(),
		Serializer:       cherrySerializer.NewProtobuf(),
		PacketListener:   nil,
		OnCreateListener: make([]cherryAgent.SessionListener, 0),
		OnCloseListener:  make([]cherryAgent.SessionListener, 0),
	}

	ws := NewTCP(address)

	return NewConnectorWithOpt(opt, ws)
}

func NewWebsocketComponent(address string) *ConnectorComponent {
	opt := cherryAgent.AgentOpt{
		Heartbeat:        60,
		DataCompression:  false,
		PacketDecoder:    cherryPacket.NewPomeloDecoder(),
		PacketEncoder:    cherryPacket.NewPomeloEncoder(),
		Serializer:       cherrySerializer.NewProtobuf(),
		PacketListener:   nil,
		OnCreateListener: make([]cherryAgent.SessionListener, 0),
		OnCloseListener:  make([]cherryAgent.SessionListener, 0),
	}

	ws := NewWebSocket(address)

	return NewConnectorWithOpt(opt, ws)
}

func NewConnectorWithOpt(opts cherryAgent.AgentOpt, connector facade.IConnector) *ConnectorComponent {
	return &ConnectorComponent{
		AgentOpt:      opts,
		connector:     connector,
		ConnectStat:   &ConnectStat{},
		packetHandles: make(map[byte]cherryAgent.PacketListener),
	}
}

func (p *ConnectorComponent) Name() string {
	return cherryConst.ConnectorPomeloComponent
}

func (p *ConnectorComponent) Init() {
	p.packetHandles[cherryPacket.Handshake] = p.handshake
	p.packetHandles[cherryPacket.HandshakeAck] = p.handshakeACK
	p.packetHandles[cherryPacket.Heartbeat] = p.heartbeat
	p.packetHandles[cherryPacket.Data] = p.handData
}

func (p *ConnectorComponent) OnAfterInit() {
	p.sessionComponent = p.App().Find(cherryConst.SessionComponent).(*cherrySession.SessionComponent)
	if p.sessionComponent == nil {
		panic("please preload session component.")
	}

	p.handlerComponent = p.App().Find(cherryConst.HandlerComponent).(*cherryHandler.HandlerComponent)
	if p.handlerComponent == nil {
		panic("preload handler component please.")
	}

	p.OnCreateSession(func(s *cherrySession.Session) (next bool, err error) {
		// increase connect stat
		p.ConnectStat.IncreaseConn()
		cherryLogger.Debugf("after create. session:%s", s)
		return true, nil
	})

	p.OnCloseSession(func(s *cherrySession.Session) (next bool, err error) {
		// decrease connect stat
		p.ConnectStat.DecreaseConn()
		return true, nil
	})

	// process packet listener
	if p.PacketListener == nil {
		p.PacketListener = func(agent *cherryAgent.Agent, packet *cherryPacket.Packet) {
			fn, found := p.packetHandles[packet.Type]
			if found == false {
				cherryLogger.Errorf("packet type not found. session = %s, packet = %s",
					agent.Session,
					packet,
				)
			}
			fn(agent, packet)
		}
	}

	// default on connect
	p.connector.OnConnect(func(conn facade.Conn) {
		// create new session
		session := p.sessionComponent.Create(cherrySession.NextSID(), p.App().NodeId())

		// create agent
		// TODO  add rpcHandler!
		agent := cherryAgent.NewAgent(p.AgentOpt, session, conn, nil)

		session.SetNetwork(agent)

		// run agent
		agent.Run()
	})

	// new goroutine
	go p.connector.OnStart()
}

func (p *ConnectorComponent) OnStop() {
	if p.sessionComponent != nil {
		p.sessionComponent.CloseAll()
	}

	if p.connector != nil {
		p.connector.OnStop()
	}
}

func (p *ConnectorComponent) OnCreateSession(listener ...cherryAgent.SessionListener) {
	p.OnCreateListener = append(p.OnCreateListener, listener...)
}

func (p *ConnectorComponent) OnCloseSession(listener ...cherryAgent.SessionListener) {
	p.OnCloseListener = append(p.OnCloseListener, listener...)
}

func (p *ConnectorComponent) SetPacketHandle(typ byte, listener cherryAgent.PacketListener) {
	p.packetHandles[typ] = listener
}

func (p *ConnectorComponent) handshake(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	data := map[string]interface{}{
		"code": 200,
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

	cherryLogger.Debugf("sid = %d request handshake", agent.Session.SID())
}

func (p *ConnectorComponent) handshakeACK(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	agent.SetStatus(cherryAgent.Working)

	cherryLogger.Debugf("sid = %d request handshakeACK", agent.Session.SID())
}

func (p *ConnectorComponent) heartbeat(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	bytes, err := agent.PacketEncoder.Encode(cherryPacket.Heartbeat, nil)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	err = agent.SendRaw(bytes)
	if err != nil {
		cherryLogger.Error(err)
	}

	cherryLogger.Debugf("sid:%d request heartbeat", agent.Session.SID())
}

func (p *ConnectorComponent) handData(agent *cherryAgent.Agent, pkg *cherryPacket.Packet) {
	if agent.Status() != cherryAgent.Working {
		cherryLogger.Warnf("status is not working. session=[%s]", agent.Session)
		return
	}

	msg, err := cherryMessage.Decode(pkg.Data)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	route, err := cherryRoute.Decode(msg.Route)
	if err != nil {
		cherryLogger.Errorf("route decode error. session=[%s] message=%s", agent.Session, msg)
		return
	}

	unHandleMessage := &cherryHandler.UnhandledMessage{
		Session: agent.Session,
		Route:   route,
		Msg:     msg,
	}

	p.handlerComponent.DoHandle(unHandleMessage)
}
