package cherryCluster

import (
	"context"
	cherryConst "github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySession "github.com/cherry-game/cherry/net/session"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"google.golang.org/grpc"
	"net"
)

var (
	GrpcOptions = []grpc.DialOption{grpc.WithInsecure()}
)

type Component struct {
	facade.Component
	mode             string
	discovery        facade.IDiscovery
	rpcServer        *grpc.Server
	clientPool       *connPool
	onCloseSessionFn func(sid facade.SID)
	onForwardFn      func(msg *cherryProto.Message)
}

func NewComponent() *Component {
	return &Component{}
}

func (c *Component) Name() string {
	return cherryConst.ClusterComponent
}

func (c *Component) Init() {
	clusterConfig := cherryProfile.Config().Get("cluster")
	if clusterConfig.LastError() != nil {
		cherryLogger.Error("`cluster` property not found in profile file.")
		return
	}

	discoveryConfig := clusterConfig.Get("discovery")
	if discoveryConfig.LastError() != nil {
		cherryLogger.Error("`discovery` property not found in profile file.")
		return
	}

	c.mode = discoveryConfig.Get("mode").ToString()
	if c.mode == "" {
		cherryLogger.Error("`discovery->mode` property not found in profile file.")
		return
	}

	c.discovery = GetDiscovery(c.mode)
	if c.discovery == nil {
		cherryLogger.Errorf("not found mode = %s property in discovery map.", c.mode)
		return
	}

	c.rpcServer = grpc.NewServer()
	c.discovery.Init(c.App(), c.rpcServer, discoveryConfig)
	c.initRPCServer()

	c.clientPool = newPool(GrpcOptions...)

}

func (c *Component) initRPCServer() {
	listener, err := net.Listen("tcp", c.App().RpcAddress())
	if err != nil {
		panic(err)
	}

	cherryProto.RegisterMemberServiceServer(c.rpcServer, c)

	go func() {
		err := c.rpcServer.Serve(listener)
		if err != nil {
			cherryLogger.Errorf("start current master server node failed: %v", err)
		}
	}()

	cherryLogger.Infof("rpc server is running. [address = %s]", c.App().RpcAddress())
}

func (c *Component) OnStop() {
	c.rpcServer.GracefulStop()
	c.clientPool.close()
	c.discovery.OnStop()
}

func (c *Component) Discovery() facade.IDiscovery {
	return c.discovery
}

func (c *Component) OnAddMember(listener facade.MemberListener) {
	c.Discovery().OnAddMember(listener)
}

func (c *Component) OnRemoveMember(listener facade.MemberListener) {
	c.Discovery().OnRemoveMember(listener)
}

func (c *Component) NewMember(_ context.Context, newMember *cherryProto.Member) (*cherryProto.Response, error) {
	c.discovery.AddMember(newMember)
	return &cherryProto.Response{}, nil
}

func (c *Component) RemoveMember(_ context.Context, node *cherryProto.NodeId) (*cherryProto.Response, error) {
	c.discovery.RemoveMember(node.Id)
	return &cherryProto.Response{}, nil
}

func (c *Component) CloseSession(_ context.Context, session *cherryProto.SessionId) (*cherryProto.Response, error) {
	if c.onCloseSessionFn != nil {
		c.onCloseSessionFn(session.Sid)
	} else {
		cherryLogger.Warnf("on close session function not found. [sessionId = %d]", session.Sid)
	}

	return &cherryProto.Response{}, nil
}

func (c *Component) Forward(_ context.Context, msg *cherryProto.Message) (*cherryProto.Response, error) {
	if c.onForwardFn != nil {
		c.onForwardFn(msg)
	} else {
		cherryLogger.Warnf("on forward function not found. [message = %v]", msg)
	}

	return &cherryProto.Response{}, nil
}

func (c *Component) SendUserMessage(session *cherrySession.Session, msg *cherryMessage.Message) {
	cherryLogger.Info("forward message to remote server ")

	// 根据 nodeType() 获取 路由的策略
	// 获取clienPool 连接信息
	// 转发消息

	//c.clientPool.GetMemberClient()
}

func (c *Component) SendSysMessage(nodeId string, msg *cherryProto.Message) {

}

// SendCloseSession move to handlerComponent
func (c *Component) SendCloseSession(sid facade.SID) {
	for _, member := range c.discovery.List() {
		if member.GetNodeId() == c.App().NodeId() {
			continue
		}

		client, found := c.clientPool.GetMemberClient(member.GetNodeId())
		if found == false {
			cherryLogger.Warnf("get member client not found. [address = %s]", member.GetAddress())
			continue
		}

		_, err := client.CloseSession(context.Background(), &cherryProto.SessionId{Sid: sid})
		if err != nil {
			cherryLogger.Warnf("[sessionId = %d] send close session fail. [error = %s]", sid, err)
			return
		}
	}
}

func (c *Component) OnCloseSession(func(sid facade.SID)) {

}

func (c *Component) OnForward(func(msg *cherryProto.Message)) {

}
