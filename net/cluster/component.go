package cherryCluster

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryNATS "github.com/cherry-game/cherry/extend/nats"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryDiscovery "github.com/cherry-game/cherry/net/discovery"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryRouter "github.com/cherry-game/cherry/net/router"
	cherrySession "github.com/cherry-game/cherry/net/session"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
)

type Component struct {
	facade.Component
	natsConfig       *cherryProfile.NatsConfig
	nats             *nats.Conn
	client           facade.RPCClient
	server           facade.RPCServer
	handlerComponent *cherryHandler.Component
}

func NewComponent(handlerComponent *cherryHandler.Component) *Component {
	return &Component{
		handlerComponent: handlerComponent,
	}
}

func (c *Component) Name() string {
	return cherryConst.ClusterComponent
}

func (c *Component) Client() facade.RPCClient {
	return c.client
}

func (c *Component) Server() facade.RPCServer {
	return c.server
}

func (c *Component) Init() {
	// init discovery
	cherryDiscovery.Init(c.App())

	c.natsConfig = cherryProfile.NewNatsConfig()

	var err error
	c.nats, err = cherryNATS.Connect(c.natsConfig)
	if err != nil {
		cherryLogger.Warnf("err = %s", err)
		return
	}

	c.client = NewNatsRPCClient(c.nats, c.natsConfig)
	c.server = NewRpcServer(c.handlerComponent, c.nats, c.client)

	c.client.Init(c.App())
	c.server.Init(c.App())
}

func (c *Component) OnStop() {
	cherryDiscovery.OnStop()
	c.client.OnStop()
	c.server.OnStop()

	c.nats.Close()
}

// ForwardLocal forward message to backend node
func (c *Component) ForwardLocal(session *cherrySession.Session, msg *cherryMessage.Message) {
	if session.IsBind() == false {
		statusCode := cherryCode.GetCodeResult(cherryCode.SessionUIDNotBind)
		session.Kick(statusCode, false)

		session.Warnf("session not bind,message forwarding is not allowed. [route = %s]", msg.Route)
		return
	}

	ctx := context.WithValue(context.Background(), cherryConst.SessionKey, session)
	member, err := cherryRouter.Route(ctx, msg.RouteInfo().NodeType(), msg)
	if member == nil || err != nil {
		session.Warnf("get node router is fail. [error = %s]", err)
		return
	}

	pbSession := &cherryProto.Session{
		Sid:        session.SID(),
		Uid:        session.UID(),
		FrontendId: session.FrontendId(),
		Ip:         session.RemoteAddress(),
		Data:       make(map[string]string),
	}

	for k, v := range session.Data() {
		pbSession.Data[k] = v
	}

	localPacket := &cherryProto.LocalPacket{
		Session: pbSession,
		MsgType: int32(msg.Type),
		MsgId:   uint32(msg.ID),
		Route:   msg.Route,
		IsError: false,
		Data:    msg.Data,
	}

	err = c.client.CallLocal(member.GetNodeId(), localPacket)
	if err != nil {
		session.Warnf("forward local fail. [error = %s]", err)
		return
	}
}
