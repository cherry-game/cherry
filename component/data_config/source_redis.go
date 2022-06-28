package cherryDataConfig

import (
	"context"
	"fmt"
	cerr "github.com/cherry-game/cherry/error"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/go-redis/redis/v8"
)

// SourceRedis redis方式获取数据配置
//
// 从profile-x.json中获取data_config的属性配置，
// 如果"data_source"的值为"redis"，则启用redis方式读取数据配置.
// 通过redis的订阅机制来触发哪个配置有变更，则进行重新加载处理.
// 程序启动后，会订阅“subscribeKey”，当有变更时，则执行加载.
type SourceRedis struct {
	changeFn     ConfigChangeFn
	close        chan struct{}
	prefixKey    string
	subscribeKey string
	address      string
	password     string
	db           int
	rdb          *redis.Client
}

func (r *SourceRedis) Name() string {
	return "redis"
}

func (r *SourceRedis) Init(_ IDataConfig) {
	r.close = make(chan struct{})

	//read data_config->file node
	redisConfig := cprofile.GetConfig("data_config").GetConfig(r.Name())
	if redisConfig.LastError() != nil {
		clog.Warnf("[data_config]->[%s] node in `%s` file not found.", r.Name(), cprofile.FileName())
		return
	}

	r.prefixKey = redisConfig.GetString("prefix_key")
	r.subscribeKey = redisConfig.GetString("subscribe_key")
	r.address = redisConfig.GetString("address")
	r.password = redisConfig.GetString("password")
	r.db = redisConfig.GetInt("db")

	r.rdb = redis.NewClient(&redis.Options{
		Addr:     r.address,
		Password: r.password,
		DB:       r.db,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			clog.Infof("data config for redis connected")
			return nil
		},
	})

	go r.newSubscribe()
}

func (r *SourceRedis) newSubscribe() {
	if r.subscribeKey == "" {
		panic("subscribe key is empty.")
	}

	sub := r.rdb.Subscribe(context.Background(), r.subscribeKey)

	defer func(sub *redis.PubSub) {
		err := sub.Close()
		if err != nil {
			clog.Warn(err)
		}
	}(sub)

	for {
		select {
		case <-r.close:
			return
		case ch := <-sub.Channel():
			if ch.Payload == "" {
				continue
			}

			clog.Infof("[name = %s] trigger file change.", ch.Payload)

			data, err := r.ReadBytes(ch.Payload)
			if err != nil {
				clog.Warnf("[name = %s] read data error = %s", ch.Payload, err)
				continue
			}

			if r.changeFn != nil {
				r.changeFn(ch.Payload, data)
			}
		}
	}
}

func (r *SourceRedis) ReadBytes(configName string) (data []byte, error error) {
	if configName == "" {
		return nil, cerr.Error("configName is empty.")
	}

	key := fmt.Sprintf("%s:%s", r.prefixKey, configName)

	return r.rdb.Get(context.Background(), key).Bytes()
}

func (r *SourceRedis) OnChange(fn ConfigChangeFn) {
	r.changeFn = fn
}

func (r *SourceRedis) Stop() {
	clog.Infof("close redis client [address = %s]", r.address)
	r.close <- struct{}{}

	if r.rdb != nil {
		err := r.rdb.Close()
		if err != nil {
			clog.Error(err)
		}
	}
}
