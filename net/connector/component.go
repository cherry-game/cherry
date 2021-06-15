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
		packetHandles    map[cherryPacket.Type]cherryAgent.PacketListener
	}
)

var (
	DefaultHeartbeat = 30 * time.Second
)

func NewTCPComponent(address string) *Component {
	opt := cherryAgent.Options{
		Heartbeat:        DefaultHeartbeat,
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

func NewWSComponent(address string) *Component {
	opt := cherryAgent.Options{
		Heartbeat:        DefaultHeartbeat,
		DataCompression:  false,
		PacketDecoder:    cherryPacket.NewPomeloDecoder(),
		PacketEncoder:    cherryPacket.NewPomeloEncoder(),
		Serializer:       cherrySerializer.NewProtobuf(),
		PacketListener:   nil,
		OnCreateListener: make([]cherryAgent.SessionListener, 0),
		OnCloseListener:  make([]cherryAgent.SessionListener, 0),
	}

	ws := NewWS(address)

	return NewConnectorWithOpt(opt, ws)
}

func NewConnectorWithOpt(opts cherryAgent.Options, connector facade.IConnector) *Component {
	return &Component{
		Options:       opts,
		connector:     connector,
		ConnectStat:   &ConnectStat{},
		packetHandles: make(map[cherryPacket.Type]cherryAgent.PacketListener),
	}
}

func (p *Component) Name() string {
	return cherryConst.ConnectorPomeloComponent
}

func (p *Component) Init() {
	p.packetHandles[cherryPacket.Handshake] = p.handshake
	p.packetHandles[cherryPacket.HandshakeAck] = p.handshakeACK
	p.packetHandles[cherryPacket.Heartbeat] = p.heartbeat
	p.packetHandles[cherryPacket.Data] = p.handData
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
		cherryLogger.Debugf("session[%s] on create.", s)
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
				cherryLogger.Errorf("session[%s], packet[%s] type not found.",
					agent.Session,
					packet,
				)
			}
			fn(agent, packet)

			// update last time
			agent.SetLastAt()
		}
	}

	// default on connect
	p.connector.OnConnect(func(conn facade.INetConn) {
		// create new session
		session := p.sessionComponent.Create(cherrySession.NextSID(), p.App().NodeId())

		// create agent
		// TODO  add rpcHandler!
		agent := cherryAgent.NewAgent(p.Options, session, conn, nil)

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
	p.OnCreateListener = append(p.OnCreateListener, listener...)
}

func (p *Component) OnCloseSession(listener ...cherryAgent.SessionListener) {
	p.OnCloseListener = append(p.OnCloseListener, listener...)
}

func (p *Component) SetPacketHandle(typ cherryPacket.Type, listener cherryAgent.PacketListener) {
	p.packetHandles[typ] = listener
}

func (p *Component) handshake(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
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

	cherryLogger.Debugf("session[%s] request handshake", agent.Session)
}

func (p *Component) handshakeACK(agent *cherryAgent.Agent, _ *cherryPacket.Packet) {
	agent.SetStatus(cherryAgent.Working)

	cherryLogger.Debugf("session[%s] request handshakeACK", agent.Session)
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
	cherryLogger.Debugf("session[%s],, request heartbeat", agent.Session)
}

func (p *Component) handData(agent *cherryAgent.Agent,
	pkg *cherryPacket.Packet) {
	if agent.Status() != cherryAgent.Working {
		cherryLogger.Warnf("status is not working. session[%s]", agent.Session)
		return
	}

	msg, err := cherryMessage.Decode(pkg.Data)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	route, err := cherryRoute.Decode(msg.Route)
	if err != nil {
		cherryLogger.Errorf("route decode error. session[%s] message[%s]", agent.Session, msg)
		return
	}

	unHandleMessage := &cherryHandler.UnhandledMessage{
		Session: agent.Session,
		Route:   route,
		Msg:     msg,
	}

	p.handlerComponent.DoHandle(unHandleMessage)
}
