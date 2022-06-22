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
		ReconnectDelay int
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

func New(config jsoniter.Any, opts ...OptionFunc) *NatsConnect {
	nats := &NatsConnect{}
	nats.loadConfig(config)

	for _, opt := range opts {
		opt(&nats.Options)
	}
	return nats
}

func (p *NatsConnect) Connect() {
	for {
		conn, err := nats.Connect(p.Address, p.GetNatsOption()...)
		if err != nil {
			cherryLogger.Warnf("nats connect fail! retrying in 3 seconds. address = %s, err = %s", p.Address, err)
			time.Sleep(3 * time.Second)
			continue
		}

		p.Conn = conn
		cherryLogger.Infof("nats is connected! [address = %s]", p.Address)
		break
	}
}

func (p *NatsConnect) loadConfig(config jsoniter.Any) {
	address := config.Get("address").ToString()
	if address == "" {
		panic("address is empty!")
	}

	p.Address = address
	p.ReconnectDelay = config.Get("reconnect_delay").ToInt()
	p.MaxReconnects = config.Get("max_reconnects").ToInt()
	p.RequestTimeout = time.Duration(config.Get("request_timeout").ToInt()) * time.Second
	p.User = config.Get("user").ToString()
	p.Password = config.Get("password").ToString()
}

func (p *NatsConnect) GetNatsOption() []nats.Option {
	var options []nats.Option

	if p.ReconnectDelay > 0 {
		options = append(options, nats.ReconnectWait(time.Duration(p.ReconnectDelay)*time.Second))
	}

	options = append(options, nats.MaxReconnects(p.MaxReconnects))

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
			cherryLogger.Infof("error = %v", nc.LastError())
		}
	}))

	options = append(options, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		cherryLogger.Warnf("isConnect = %v. %s on connection for subscription on %q",
			nc.IsConnected(),
			err.Error(),
			sub.Subject,
		)
	}))

	if p.User != "" {
		options = append(options, nats.UserInfo(p.User, p.Password))
	}

	return options
}

func WithAddress(address string) OptionFunc {
	return func(opts *Options) {
		opts.Address = address
	}
}

func WithParams(reconnectDelay int, maxReconnects int, requestTimeout time.Duration) OptionFunc {
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
