package cherryNats

import (
	"sync/atomic"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

var (
	connectPool []*Connect               // connect pool
	roundIndex  *uint64    = new(uint64) // round-robin index
)

func NewConnectPool(replySubject string, config cfacade.ProfileJSON, isConnect bool) {
	resetConnectPool()

	var (
		address        = config.GetString("address")
		user           = config.GetString("user")
		pwd            = config.GetString("password")
		maxReconnects  = config.GetInt("max_reconnects")
		poolSize       = config.GetInt("pool_size", 1)
		isStats        = config.GetBool("is_stats")
		statsInterval  = config.GetInt("stats_interval", 30)
		reconnectDelay = config.GetDuration("reconnect_delay", 1) * time.Second
		requestTimeout = config.GetDuration("request_timeout", 1) * time.Second
	)

	for id := 1; id <= poolSize; id++ {
		conn := NewConnect(id, replySubject,
			WithAddress(address),
			WithReconnectDelay(reconnectDelay),
			WithRequestTimeout(requestTimeout),
			WithAuth(user, pwd),
			WithParams(maxReconnects),
			WithIsStats(isStats),
			WithStatsInterval(statsInterval),
		)

		connectPool = append(connectPool, conn)
	}

	if isConnect {
		for _, conn := range connectPool {
			conn.Connect()
		}

		clog.Infof("Nats has connected! [poolSize = %d]", poolSize)
	}
}

func GetConnectPool() []*Connect {
	return connectPool
}

func GetConnect() *Connect {
	size := connectPoolSize()
	if size == 0 {
		return nil
	}

	index := atomic.AddUint64(roundIndex, 1)
	return connectPool[index%size]
}

func CloseConnectPool() {
	resetConnectPool()
}

func resetConnectPool() {
	if len(connectPool) > 0 {
		for _, conn := range connectPool {
			conn.Close()
		}
		clog.Infof("Nats connect pool execute Close() [connectSize = %d]", connectPoolSize())
	}

	connectPool = nil
	atomic.StoreUint64(roundIndex, 0)
}

func connectPoolSize() uint64 {
	return uint64(len(connectPool))
}
