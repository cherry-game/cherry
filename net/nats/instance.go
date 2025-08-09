package cherryNats

import (
	"sync/atomic"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

var (
	connectPool []*Connect // connect pool
	connectSize uint64     // connect size
	roundIndex  *uint64    // round-robin index
)

func NewConnectPool(config cfacade.ProfileJSON, isConnect bool) {
	poolSize := config.GetInt("pool_size", 1)

	for i := range poolSize {
		connectSize += 1

		conn := NewFromConfig(config)
		connectPool = append(connectPool, conn)
		clog.Infof("[%d] Add nats client", i)
	}

	if isConnect {
		for _, conn := range connectPool {
			conn.Connect()
		}
	}
}

func GetConnectPool() []*Connect {
	return connectPool
}

func GetConnect() *Connect {
	index := atomic.AddUint64(roundIndex, 1)
	return connectPool[index%connectSize]
}

func ConnectClose() {
	for _, conn := range connectPool {
		conn.Close()
	}
}

func NewFromConfig(config cfacade.ProfileJSON) *Connect {
	conn := New()
	conn.address = config.GetString("address")
	conn.maxReconnects = config.GetInt("max_reconnects")
	conn.reconnectDelay = config.GetDuration("reconnect_delay", 1) * time.Second
	conn.requestTimeout = config.GetDuration("request_timeout", 1) * time.Second
	conn.user = config.GetString("user")
	conn.password = config.GetString("password")

	if conn.address == "" {
		panic("address is empty!")
	}

	return conn
}
