package cherryNats

import (
	"sync/atomic"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
)

var (
	connectPool    []*Connect                      // connect pool
	connectSize    uint64                          // connect size
	roundIndex     *uint64       = new(uint64)     // round-robin index
	reconnectDelay time.Duration = 1 * time.Second // reconnect delay
	requestTimeout time.Duration = 2 * time.Second // request timeout
)

func NewConnectPool(config cfacade.ProfileJSON, isConnect bool) {
	reconnectDelay = config.GetDuration("reconnect_delay", 1) * time.Second
	requestTimeout = config.GetDuration("request_timeout", 1) * time.Second

	poolSize := config.GetInt("pool_size", 1)
	for i := 1; i <= poolSize; i++ {
		connectSize += 1
		conn := NewFromConfig(i, config)
		connectPool = append(connectPool, conn)
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

func ReconnectDelay() time.Duration {
	return reconnectDelay
}

func NewFromConfig(index int, config cfacade.ProfileJSON) *Connect {
	conn := New()
	conn.index = index
	conn.address = config.GetString("address")
	conn.maxReconnects = config.GetInt("max_reconnects")
	conn.user = config.GetString("user")
	conn.password = config.GetString("password")

	if conn.address == "" {
		panic("address is empty!")
	}

	return conn
}
