package cherryCluster

import (
	"context"
	cherryMap "github.com/cherry-game/cherry/extend/map"
	cherryFacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	"google.golang.org/grpc"
	"time"
)

type (
	connPool struct {
		opts  []grpc.DialOption
		pools *cherryMap.SafeMap
	}

	conn struct {
		nodeId        string
		address       string
		rpcClientConn *grpc.ClientConn
		masterClient  cherryProto.MasterServiceClient
		memberClient  cherryProto.MemberServiceClient
	}
)

func newPool(opts ...grpc.DialOption) *connPool {
	c := &connPool{
		opts:  opts,
		pools: cherryMap.NewSafeMap(),
	}
	return c
}

func (c *connPool) GetMemberClient(nodeId string) (cherryProto.MemberServiceClient, bool) {
	clientConn, found := c.getConn(nodeId)
	if found == false {
		return nil, found
	}

	return clientConn.memberClient, true
}

func (c *connPool) GetMasterClient(nodeId string) (cherryProto.MasterServiceClient, bool) {
	clientConn, found := c.getConn(nodeId)
	if found == false {
		return nil, found
	}

	return clientConn.masterClient, true
}

func (c *connPool) addConn(member cherryFacade.IMember) {
	value := c.pools.Get(member.GetNodeId())
	if value != nil {
		clientConn := value.(*conn)
		if clientConn.address == member.GetAddress() {
			return
		}
		c.remove(member.GetNodeId())
	}

	rpcClientConn, err := c.buildRpcClientConn(member.GetAddress())
	if err != nil {
		cherryLogger.Warnf("build client conn error. [error = %s]", err)
		return
	}

	newConn := &conn{
		nodeId:        member.GetNodeId(),
		address:       member.GetAddress(),
		rpcClientConn: rpcClientConn,
		masterClient:  cherryProto.NewMasterServiceClient(rpcClientConn),
		memberClient:  cherryProto.NewMemberServiceClient(rpcClientConn),
	}

	c.pools.Set(member.GetNodeId(), newConn)
}

func (c *connPool) remove(nodeId string) {
	newConn, found := c.getConn(nodeId)
	if found == false {
		return
	}

	newConn.disconnect()
	c.pools.Delete(nodeId)
}

func (c *connPool) getConn(nodeId string) (*conn, bool) {
	value := c.pools.Get(nodeId)
	if value != nil {
		return value.(*conn), true
	}

	return nil, false
}

func (c *connPool) buildRpcClientConn(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	dialConn, err := grpc.DialContext(
		ctx,
		address,
		c.opts...,
	)
	cancel()

	if err != nil {
		return nil, err
	}

	return dialConn, nil
}

func (c *conn) disconnect() {
	if c.rpcClientConn != nil {
		err := c.rpcClientConn.Close()
		if err != nil {
			cherryLogger.Error(err)
		}
	}
}

func (c *connPool) close() {
	for key, value := range c.pools.Items() {
		con := value.(*conn)
		con.disconnect()
		c.pools.Delete(key)
	}

	cherryLogger.Infof("conn pool is closed.")
}
