package cherryProfile

import (
	jsoniter "github.com/json-iterator/go"
	"time"
)

type NatsConfig struct {
	Address        string
	ReconnectDelay time.Duration
	MaxReconnects  int
	RequestTimeout time.Duration
	User           string
	Password       string
}

func NewNatsConfig() *NatsConfig {
	cluster := Config().Get("cluster")
	nats := cluster.Get("nats")
	return NewNatsWithConfig(nats)
}

func NewNatsWithConfig(config jsoniter.Any) *NatsConfig {
	address := config.Get("address").ToString()
	if address == "" {
		panic("address is blank")
	}

	reconnectDelay := config.Get("reconnect_delay").ToInt64()
	if reconnectDelay < 1 {
		reconnectDelay = 1
	}

	maxReconnects := config.Get("max_reconnects").ToInt()
	if maxReconnects < 1 {
		maxReconnects = 10
	}

	requestTimeout := config.Get("request_timeout").ToInt64()
	if requestTimeout < 1 {
		requestTimeout = 1
	}

	user := config.Get("user").ToString()
	password := config.Get("password").ToString()

	cfg := &NatsConfig{
		Address:        address,
		ReconnectDelay: time.Duration(reconnectDelay) * time.Second,
		MaxReconnects:  maxReconnects,
		RequestTimeout: time.Duration(requestTimeout) * time.Second,
		User:           user,
		Password:       password,
	}

	return cfg
}