package cherryNats

import (
	cherryLogger "github.com/cherry-game/cherry/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"time"
)

type (
	OptionFunc func(opts *Options)

	Options struct {
		Address        string
		ReconnectDelay time.Duration
		MaxReconnects  int
		RequestTimeout time.Duration
		User           string
		Password       string
	}

	NatsConnect struct {
		*nats.Conn
		Options
	}
)

func (p *NatsConnect) Connect() {
	for {
		conn, err := nats.Connect(p.Address, p.GetNatsOption()...)
		if err != nil {
			cherryLogger.Warnf("nats connect fail! waiting... err = %s", p.Address, err)
			time.Sleep(3 * time.Second)
			continue
		}

		p.Conn = conn
		cherryLogger.Infof("nats is connected! [address = %s]", p.Address)
		break
	}
}

func NewNats(opts ...OptionFunc) *NatsConnect {
	nats := &NatsConnect{}

	for _, opt := range opts {
		opt(&nats.Options)
	}

	return nats
}

func (p *NatsConnect) loadConfig(config jsoniter.Any) {
	address := config.Get("address").ToString()
	if address == "" {
		panic("address is empty!")
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

	p.Address = address
	p.ReconnectDelay = time.Duration(reconnectDelay) * time.Second
	p.MaxReconnects = maxReconnects
	p.RequestTimeout = time.Duration(requestTimeout) * time.Second
	p.User = config.Get("user").ToString()
	p.Password = config.Get("password").ToString()
}

func (p *NatsConnect) GetNatsOption() []nats.Option {
	var options []nats.Option

	if p.ReconnectDelay > 0 {
		options = append(options, nats.ReconnectWait(p.ReconnectDelay))
	}

	if p.MaxReconnects > 0 {
		options = append(options, nats.MaxReconnects(p.MaxReconnects))
	}

	options = append(options, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			cherryLogger.Warnf("disconnect error. [error = %v]", err)
		}
	}))

	options = append(options, nats.ReconnectHandler(func(nc *nats.Conn) {
		cherryLogger.Warnf("reconnected [%s]", nc.ConnectedUrl())
	}))

	options = append(options, nats.ClosedHandler(func(nc *nats.Conn) {
		cherryLogger.Infof("exiting... %s", p.Address)
		if nc.LastError() != nil {
			cherryLogger.Infof(".[error = %v]", nc.LastError())
		}
	}))

	if p.User != "" {
		options = append(options, nats.UserInfo(p.User, p.Password))
	}

	return options
}

func WithAdreess(address string) OptionFunc {
	return func(opts *Options) {
		opts.Address = address
	}
}

func WithParamsAdreess(reconnectDelay time.Duration, maxReconnects int, requestTimeout time.Duration) OptionFunc {
	return func(opts *Options) {
		opts.ReconnectDelay = reconnectDelay
		opts.MaxReconnects = maxReconnects
		opts.RequestTimeout = requestTimeout
	}
}

func WithAuth(user, password string) OptionFunc {
	return func(opts *Options) {
		opts.User = user
		opts.Password = password
	}
}
