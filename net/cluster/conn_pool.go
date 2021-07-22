package cherryCluster

import (
	"context"
	cherryFacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type (
	connPool struct {
		sync.RWMutex
		opts  []grpc.DialOption
		pools sync.Map
	}

	conn struct {
		nodeId       string
		address      string
		clientConn   *grpc.ClientConn
		masterClient cherryProto.MasterServiceClient
		memberClient cherryProto.MemberServiceClient
	}
)

func newPool(opts ...grpc.DialOption) *connPool {
	c := &connPool{
		opts: opts,
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

func (c *connPool) add(member cherryFacade.IMember) {
	if value, found := c.pools.Load(member.GetNodeId()); found {
		clientConn := value.(*conn)
		if clientConn.address == member.GetAddress() {
			return
		}

		c.remove(member.GetNodeId())
	}

	clientConn, err := c.buildClientConn(member.GetAddress())
	if err != nil {
		cherryLogger.Warnf("build client conn error. [error = %s]", err)
		return
	}

	newConn := &conn{
		nodeId:       member.GetNodeId(),
		address:      member.GetAddress(),
		clientConn:   clientConn,
		masterClient: cherryProto.NewMasterServiceClient(clientConn),
		memberClient: cherryProto.NewMemberServiceClient(clientConn),
	}

	c.pools.Store(member.GetNodeId(), newConn)
}

func (c *connPool) remove(nodeId string) {
	clientConn, found := c.getConn(nodeId)
	if found == false {
		return
	}

	clientConn.disconnect()
	c.pools.Delete(nodeId)
}

func (c *connPool) getConn(nodeId string) (*conn, bool) {
	if clientConn, found := c.pools.Load(nodeId); found {
		return clientConn.(*conn), true
	}

	return nil, false
}

func (c *connPool) buildClientConn(address string) (*grpc.ClientConn, error) {
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
	if c.clientConn != nil {
		err := c.clientConn.Close()
		if err != nil {
			cherryLogger.Error(err)
		}
	}
}

func (c *connPool) close() {
	c.RLock()
	defer c.RUnlock()

	c.pools.Range(func(key, value interface{}) bool {
		con := value.(*conn)
		con.disconnect()
		c.pools.Delete(key)
		return true
	})

	cherryLogger.Infof("conn pool is closed.")
}
