package cherryDataConfig

import (
	"context"
	"fmt"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
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
	config := cherryProfile.GetConfig("data_config")

	redisNode := config.Get(r.Name())
	if redisNode == nil {
		cherryLogger.Warnf("[data_config]->[%s] node in `%s` file not found.", r.Name(), cherryProfile.FileName())
		return
	}

	r.prefixKey = redisNode.Get("prefix_key").ToString()
	r.subscribeKey = redisNode.Get("subscribe_key").ToString()
	r.address = redisNode.Get("address").ToString()
	r.password = redisNode.Get("password").ToString()
	r.db = redisNode.Get("db").ToInt()

	r.rdb = redis.NewClient(&redis.Options{
		Addr:     r.address,
		Password: r.password,
		DB:       r.db,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			cherryLogger.Infof("data config for redis connected")
			return nil
		},
	})

	go r.newSubscribe()
}

func (r *SourceRedis) newSubscribe() {
	sub := r.rdb.Subscribe(context.Background(), r.subscribeKey)

	defer func(sub *redis.PubSub) {
		err := sub.Close()
		if err != nil {
			cherryLogger.Warn(err)
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

			cherryLogger.Infof("[name = %s] trigger file change.", ch.Payload)

			data, err := r.ReadBytes(ch.Payload)
			if err != nil {
				cherryLogger.Warnf("[name = %s] read data error = %s", ch.Payload, err)
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
		return nil, cherryError.Error("configName is empty.")
	}

	key := fmt.Sprintf("%s:%s", r.prefixKey, configName)

	return r.rdb.Get(context.Background(), key).Bytes()
}

func (r *SourceRedis) OnChange(fn ConfigChangeFn) {
	r.changeFn = fn
}

func (r *SourceRedis) Stop() {
	cherryLogger.Infof("close redis client [address = %s]", r.address)
	r.close <- struct{}{}

	if r.rdb != nil {
		err := r.rdb.Close()
		if err != nil {
			cherryLogger.Error(err)
		}
	}
}
