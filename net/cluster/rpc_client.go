package cherryCluster

import (
	"context"
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/logger"
	"google.golang.org/grpc"
	"sync"
	"sync/atomic"
	"time"
)

var (
	GrpcOptions = []grpc.DialOption{grpc.WithInsecure()}
)

type (
	rpcClient struct {
		sync.RWMutex
		isClosed bool
		pools    map[string]*connPool
	}

	connPool struct {
		index    uint32
		connects []*grpc.ClientConn
	}
)

func newConnArray(maxSize uint, addr string) (*connPool, error) {
	a := &connPool{
		index:    0,
		connects: make([]*grpc.ClientConn, maxSize),
	}

	if err := a.init(addr); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *connPool) init(addr string) error {
	for i := range a.connects {
		//2秒后重试
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		//拨号创建连接
		conn, err := grpc.DialContext(
			ctx,
			addr,
			GrpcOptions...,
		)

		cancel()
		if err != nil {
			// Cleanup if the initialization fails.
			a.Close()
			return err
		}

		a.connects[i] = conn
	}
	return nil
}

func (a *connPool) Get() *grpc.ClientConn {
	// 每调用一次index+1，与 当前连接数取模，这样起到轮循的效果
	// 这也可能导致问题，同一个角色的消息因为hash到不同的连接对象，结果后发送的数据先到了目标，造成顺序问题
	next := atomic.AddUint32(&a.index, 1) % uint32(len(a.connects))
	return a.connects[next]
}

func (a *connPool) Close() {
	for i, c := range a.connects {
		if c != nil {
			err := c.Close()
			if err != nil {
				cherryLogger.Warn(err)
			}
			a.connects[i] = nil
		}
	}
}

func newRPCClient() *rpcClient {
	return &rpcClient{
		pools: make(map[string]*connPool),
	}
}

func (c *rpcClient) getConnPool(addr string) (*connPool, error) {
	c.RLock()

	if c.isClosed {
		c.RUnlock()
		return nil, cherryError.Error("rpc client is closed")
	}

	array, ok := c.pools[addr]
	c.RUnlock()

	if !ok {
		var err error
		array, err = c.createConnPool(addr)
		if err != nil {
			return nil, err
		}
	}
	return array, nil
}

func (c *rpcClient) createConnPool(addr string) (*connPool, error) {
	c.Lock()
	defer c.Unlock()
	array, ok := c.pools[addr]
	if !ok {
		var err error
		// TODO: make conn count configurable
		//池里10个连接
		array, err = newConnArray(10, addr)
		if err != nil {
			return nil, err
		}
		c.pools[addr] = array
	}
	return array, nil
}

func (c *rpcClient) closePool() {
	c.Lock()
	if !c.isClosed {
		c.isClosed = true
		// close all connections
		for _, array := range c.pools {
			array.Close()
		}
	}
	c.Unlock()
}
