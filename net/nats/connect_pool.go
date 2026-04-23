package cherryNats

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/nats-io/nats.go"
)

var (
	defaultPool     = &ConnectPool{}
	defaultPoolOnce sync.Once
)

func InitPool(replySubject string, config cfacade.ProfileJSON, isConnect bool) *ConnectPool {
	defaultPoolOnce.Do(func() {
		defaultPool = NewConnectPool(replySubject, config, isConnect)
	})
	return defaultPool
}

func Pool() *ConnectPool {
	return defaultPool
}

type ConnectPool struct {
	connects   []*Connect
	roundIndex uint64
}

func NewConnectPool(replySubject string, config cfacade.ProfileJSON, isConnect bool) *ConnectPool {
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
		pool           = &ConnectPool{}
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

		pool.connects = append(pool.connects, conn)
	}

	if isConnect {
		for _, conn := range pool.connects {
			conn.Connect()
		}

		clog.Infof("Nats has connected! [poolSize = %d]", poolSize)
	}

	return pool
}

func (p *ConnectPool) GetConnect() *Connect {
	size := len(p.connects)
	if size == 0 {
		return nil
	}

	index := atomic.AddUint64(&p.roundIndex, 1)
	return p.connects[index%uint64(size)]
}

func PoolClose() {
	for _, conn := range defaultPool.connects {
		if conn != nil {
			conn.Close()
		}
	}

	clog.Infof("Nats connect pool closed. [connectSize = %d]", len(defaultPool.connects))
}

func Publish(subject string, data []byte) error {
	c := defaultPool.GetConnect()
	if c == nil {
		return fmt.Errorf("nats connect pool is nil")
	}

	return c.Publish(subject, data)
}

func PublishMsg(msg *nats.Msg) error {
	c := defaultPool.GetConnect()
	if c == nil {
		return fmt.Errorf("nats connect pool is nil")
	}
	return c.PublishMsg(msg)
}

func Request(subject string, data []byte, timeout ...time.Duration) ([]byte, error) {
	c := defaultPool.GetConnect()
	if c == nil {
		return nil, fmt.Errorf("nats connect pool is nil")
	}
	return c.Request(subject, data, timeout...)
}

func RequestSync(reqID, subject string, data []byte, timeout ...time.Duration) ([]byte, error) {
	c := defaultPool.GetConnect()
	if c == nil {
		return nil, fmt.Errorf("nats connect pool is nil")
	}
	return c.RequestSync(reqID, subject, data, timeout...)
}

func ReplySync(reqID, reply string, data []byte) error {
	c := defaultPool.GetConnect()
	if c == nil {
		return fmt.Errorf("nats connect pool is nil")
	}
	return c.ReplySync(reqID, reply, data)
}

func Subscribe(subject string, cb nats.MsgHandler) error {
	c := defaultPool.GetConnect()
	if c == nil {
		return fmt.Errorf("nats connect pool is nil")
	}
	return c.Subscribe(subject, cb)
}
