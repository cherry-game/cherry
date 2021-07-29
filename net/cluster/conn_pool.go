package cherryCluster

import (
	cherryMap "github.com/cherry-game/cherry/extend/map"
	cherryFacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	"google.golang.org/grpc"
	"sync"
)

type (
	connPool struct {
		opts  []grpc.DialOption
		pools *cherryMap.SafeMap
	}

	clientConn struct {
		sync.Mutex
		nodeId        string
		address       string
		connected     bool
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
	conn, found := c.getConn(nodeId)
	if found == false {
		return nil, found
	}

	if err := conn.connect(); err != nil {
		cherryLogger.Warnf("[grpc client] unable to connect to server %s at %s: %v", nodeId, conn.address, err)
		return nil, false
	}

	return conn.memberClient, true
}

func (c *connPool) GetMasterClient(nodeId string) (cherryProto.MasterServiceClient, bool) {
	conn, found := c.getConn(nodeId)
	if found == false {
		return nil, found
	}

	if err := conn.connect(); err != nil {
		cherryLogger.Warnf("[grpc client] unable to connect to server %s at %s: %v", nodeId, conn.address, err)
		return nil, false
	}

	return conn.masterClient, true
}

func (c *connPool) addConn(member cherryFacade.IMember) {
	value := c.pools.Get(member.GetNodeId())
	if value != nil {
		c.remove(member.GetNodeId())
	}

	newConn := &clientConn{
		nodeId:    member.GetNodeId(),
		address:   member.GetAddress(),
		connected: false,
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

func (c *connPool) getConn(nodeId string) (*clientConn, bool) {
	value := c.pools.Get(nodeId)
	if value != nil {
		return value.(*clientConn), true
	}

	return nil, false
}

func (c *clientConn) connect() error {
	c.Lock()
	defer c.Unlock()

	if c.connected {
		return nil
	}

	rpcClientConn, err := grpc.Dial(
		c.address,
		grpc.WithInsecure(),
	)

	if err != nil {
		return err
	}

	c.masterClient = cherryProto.NewMasterServiceClient(rpcClientConn)
	c.memberClient = cherryProto.NewMemberServiceClient(rpcClientConn)
	c.rpcClientConn = rpcClientConn
	c.connected = true

	return nil
}

func (c *clientConn) disconnect() {
	c.Lock()
	c.connected = false
	c.Unlock()

	if c.rpcClientConn != nil {
		err := c.rpcClientConn.Close()
		if err != nil {
			cherryLogger.Error(err)
		}
	}
}

func (c *connPool) close() {
	for key, value := range c.pools.Items() {
		con := value.(*clientConn)
		con.disconnect()
		c.pools.Delete(key)
	}

	cherryLogger.Infof("clientConn pool is closed.")
}
