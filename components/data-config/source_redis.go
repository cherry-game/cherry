package cherryDataConfig

import (
	"context"
	"fmt"

	cerr "github.com/cherry-game/cherry/error"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/go-redis/redis/v8"
)

type (
	// SourceRedis redis方式获取数据配置
	//
	// 从profile-x.json中获取data_config的属性配置，
	// 如果"data_source"的值为"redis"，则启用redis方式读取数据配置.
	// 通过redis的订阅机制来触发哪个配置有变更，则进行重新加载处理.
	// 程序启动后，会订阅“subscribeKey”，当有变更时，则执行加载.
	SourceRedis struct {
		redisConfig
		changeFn ConfigChangeFn
		close    chan bool
		rdb      *redis.Client
	}

	redisConfig struct {
		Address      string `json:"address"`       // redis地址
		Password     string `json:"password"`      // 密码
		DB           int    `json:"db"`            // db index
		PrefixKey    string `json:"prefix_key"`    // 前缀
		SubscribeKey string `json:"subscribe_key"` // 订阅key
	}
)

func (r *SourceRedis) Name() string {
	return "redis"
}

func (r *SourceRedis) Init(_ IDataConfig) {
	//read data_config->file node
	dataConfig := cprofile.GetConfig("data_config").GetConfig(r.Name())
	if dataConfig.Unmarshal(&r.redisConfig) != nil {
		clog.Warnf("[data_config]->[%s] node in `%s` file not found.", r.Name(), cprofile.Name())
		return
	}

	r.newRedis()
	r.close = make(chan bool)

	go r.newSubscribe()
}

func (r *SourceRedis) newRedis() {
	r.rdb = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.DB,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			clog.Infof("data config for redis connected")
			return nil
		},
	})
}

func (r *SourceRedis) newSubscribe() {
	if r.SubscribeKey == "" {
		panic("subscribe key is empty.")
	}

	sub := r.rdb.Subscribe(context.Background(), r.SubscribeKey)

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

func (r *SourceRedis) ReadBytes(configName string) ([]byte, error) {
	if configName == "" {
		return nil, cerr.Error("configName is empty.")
	}

	key := fmt.Sprintf("%s:%s", r.PrefixKey, configName)

	return r.rdb.Get(context.Background(), key).Bytes()
}

func (r *SourceRedis) OnChange(fn ConfigChangeFn) {
	r.changeFn = fn
}

func (r *SourceRedis) Stop() {
	clog.Infof("close redis client [address = %s]", r.Address)
	r.close <- true

	if r.rdb != nil {
		err := r.rdb.Close()
		if err != nil {
			clog.Error(err)
		}
	}
}
