package cherryCluster

import (
	"context"
	cherryConst "github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"google.golang.org/grpc"
	"net"
)

var (
	GrpcOptions = []grpc.DialOption{grpc.WithInsecure()}
)

type Component struct {
	facade.Component
	mode       string
	discovery  facade.IDiscovery
	rpcServer  *grpc.Server
	clientPool *connPool
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

func (c *Component) CloseSession(_ context.Context, _ *cherryProto.SessionId) (*cherryProto.Response, error) {
	// 获取sessionComponent
	// 移除session
	// 发布一个 session close 事件
	return &cherryProto.Response{}, nil
}

func (c *Component) Forward(_ context.Context, _ *cherryProto.Message) (*cherryProto.Response, error) {
	return &cherryProto.Response{}, nil
}

// SendCloseSession move to handlerComponent
func (c *Component) SendCloseSession(sessionId facade.SID) error {
	for _, member := range c.discovery.List() {
		if member.GetNodeId() == c.App().NodeId() {
			continue
		}

		client, found := c.clientPool.GetMemberClient(member.GetNodeId())
		if found == false {
			cherryLogger.Warnf("get member client not found. [address = %s]", member.GetAddress())
			continue
		}

		_, err := client.CloseSession(context.Background(), &cherryProto.SessionId{Sid: sessionId})
		if err != nil {
			return err
		}
	}

	return nil
}
