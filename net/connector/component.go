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
	"github.com/cherry-game/cherry/net/session"
	"time"
)

type (
	// Component (连接器组件适用于前端节点)
	Component struct {
		facade.Component
		cherryAgent.Options
		ConnectStat      *ConnectStat
		connector        facade.IConnector
		sessionComponent cherrySession.IComponent
		handlerComponent cherryHandler.IComponent
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
		ConnectStat: &ConnectStat{},
		connector:   connector,
	}
}

func (c *Component) Name() string {
	return cherryConst.ConnectorPomeloComponent
}

func (c *Component) Init() {

}

func (c *Component) OnAfterInit() {
	var found = false
	c.sessionComponent, found = c.App().Find(cherryConst.SessionComponent).(cherrySession.IComponent)
	if found == false {
		panic("session component must be preloaded.")
	}

	c.handlerComponent, found = c.App().Find(cherryConst.HandlerComponent).(cherryHandler.IComponent)
	if found == false {
		panic("handler component must be preloaded.")
	}

	// set rpc handler
	c.Options.RPCHandler = c.handlerComponent.PostMessage

	c.sessionComponent.AddOnCreate(func(session *cherrySession.Session) (next bool) {
		c.ConnectStat.IncreaseConn()
		session.Debugf("session on create. address[%s], state[%s]", session.RemoteAddress(), c.ConnectStat.PrintInfo())
		c.ConnectStat.PrintInfo()
		return true
	})

	c.sessionComponent.AddOnClose(func(session *cherrySession.Session) (next bool) {
		c.ConnectStat.DecreaseConn()
		session.Debugf("session on closed. address[%s], state[%s]", session.RemoteAddress(), c.ConnectStat.PrintInfo())
		return true
	})

	//add packet listener
	c.AddPacketHandle(cherryPacket.Handshake, c.handshake)
	c.AddPacketHandle(cherryPacket.HandshakeAck, c.handshakeACK)
	c.AddPacketHandle(cherryPacket.Heartbeat, c.heartbeat)
	c.AddPacketHandle(cherryPacket.Data, c.handData)

	c.connector.OnConnect(func(conn facade.INetConn) {
		// create agent
		agent := cherryAgent.NewAgent(c.App(), c.Options, conn)
		// create new session
		session := c.sessionComponent.Create(agent)
		agent.Session = session

		// run agent
		agent.Run()
	})

	// new goroutine for connector
	go c.connector.OnStart()
}

func (c *Component) OnStop() {
	if c.sessionComponent != nil {
		c.sessionComponent.CloseAll()
	}

	if c.connector != nil {
		c.connector.OnStop()
	}
}

func (c *Component) AddPacketHandle(typ cherryPacket.Type, listener cherryAgent.PacketListener) {
	c.Options.PacketListener[typ] = listener
}

func (c *Component) handshake(agent *cherryAgent.Agent, _ facade.IPacket) {
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

	bytes, err := c.App().PacketEncode(cherryPacket.Handshake, jsonData)
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

func (c *Component) handshakeACK(agent *cherryAgent.Agent, _ facade.IPacket) {
	agent.SetStatus(cherryAgent.Working)
	agent.Session.Debug("request handshakeACK.")
}

func (c *Component) heartbeat(agent *cherryAgent.Agent, _ facade.IPacket) {
	bytes, err := c.App().PacketEncode(cherryPacket.Heartbeat, nil)
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

func (c *Component) handData(agent *cherryAgent.Agent, pkg facade.IPacket) {
	if agent.Status() != cherryAgent.Working {
		agent.Session.Warnf("status is not working. status[%d]", agent.Status())
		return
	}

	msg, err := cherryMessage.Decode(pkg.Data())
	if err != nil {
		agent.Session.Warnf("packet decode error. data[%s], error[%s].", pkg.Data(), err)
		return
	}

	c.handlerComponent.PostMessage(agent.Session, msg)
}
