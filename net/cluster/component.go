package cherryCluster

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryConst "github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	cherryDiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cherryNats "github.com/cherry-game/cherry/net/cluster/nats"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryRouter "github.com/cherry-game/cherry/net/router"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type Component struct {
	facade.Component
	client facade.RPCClient
	server facade.RPCServer
}

func NewComponent() *Component {
	return &Component{}
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
	cherryNats.Init()

	c.client = NewRPCClient()
	c.client.Init(c.App())

	c.server = NewRPCServer(c.client)
	c.server.Init(c.App())

	// init discovery
	cherryDiscovery.Init(c.App())
}

func (c *Component) OnStop() {
	cherryDiscovery.OnStop()
	c.client.OnStop()
	c.server.OnStop()
	cherryNats.Conn().Close()
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
