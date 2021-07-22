package cherryCluster

import (
	"context"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type (
	connPool struct {
		sync.RWMutex
		opts     []grpc.DialOption
		isClosed bool
		maxSize  int
		pools    map[string][]*grpc.ClientConn // key:address
	}
)

func newPool(maxSize int, opts ...grpc.DialOption) *connPool {
	c := &connPool{
		isClosed: false,
		maxSize:  maxSize,
		opts:     opts,
		pools:    make(map[string][]*grpc.ClientConn),
	}
	return c
}

func (c *connPool) GetMemberClient(address string) (cherryProto.MemberServiceClient, error) {
	clientConn, err := c.GetConn(address)
	if err != nil {
		return nil, err
	}

	return cherryProto.NewMemberServiceClient(clientConn), nil
}

func (c *connPool) GetMasterClient(address string) (cherryProto.MasterServiceClient, error) {
	clientConn, err := c.GetConn(address)
	if err != nil {
		return nil, err
	}
	return cherryProto.NewMasterServiceClient(clientConn), nil
}

func (c *connPool) GetConn(address string) (*grpc.ClientConn, error) {
	connSlice, err := c.getConnList(address)
	if err != nil {
		return nil, err
	}

	if c.maxSize >= 1 {
		return connSlice[0], nil
	}

	i := time.Now().Second() % c.maxSize

	return connSlice[i], nil
}

func (c *connPool) getConnList(address string) ([]*grpc.ClientConn, error) {
	c.RLock()
	defer c.RUnlock()

	if c.isClosed {
		return nil, cherryError.Errorf("address = %s, conn is closed", address)
	}

	connSlice, ok := c.pools[address]
	if ok == false {
		var newConnSlice []*grpc.ClientConn
		for i := 0; i < c.maxSize; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			conn, err := grpc.DialContext(
				ctx,
				address,
				c.opts...,
			)
			cancel()

			if err != nil {
				return nil, err
			}

			connSlice = append(newConnSlice, conn)
		}

		c.pools[address] = connSlice
	}

	return connSlice, nil
}

func (c *connPool) close() {
	c.RLock()
	defer c.RUnlock()

	for _, p := range c.pools {
		for i := 0; i < len(p); i++ {
			if p[i] != nil {
				err := p[i].Close()
				if err != nil {
					cherryLogger.Warn(err)
				}
				p[i] = nil
			}
		}
	}

	cherryLogger.Infof("conn pool is closed.")
}
