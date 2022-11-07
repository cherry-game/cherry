package cherryNats

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
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

func New(opts ...OptionFunc) *NatsConnect {
	conn := &NatsConnect{}
	for _, opt := range opts {
		opt(&conn.Options)
	}
	return conn
}

func (p *NatsConnect) Connect() {
	for {
		conn, err := nats.Connect(p.Address, p.GetNatsOption()...)
		if err != nil {
			clog.Warnf("nats connect fail! retrying in 3 seconds. err = %s", err)
			time.Sleep(3 * time.Second)
			continue
		}

		p.Conn = conn
		clog.Infof("nats is connected! [address = %s]", p.Address)
		break
	}
}

func (p *NatsConnect) LoadConfig(config cfacade.JsonConfig) {
	address := config.Get("address").ToString()
	if address == "" {
		panic("address is empty!")
	}

	p.Address = address
	p.ReconnectDelay = config.GetInt("reconnect_delay")
	p.MaxReconnects = config.GetInt("max_reconnects")
	p.RequestTimeout = time.Duration(config.GetInt("request_timeout")) * time.Second
	p.User = config.GetString("user")
	p.Password = config.GetString("password")
}

func (p *NatsConnect) GetNatsOption() []nats.Option {
	var options []nats.Option

	if p.ReconnectDelay > 0 {
		options = append(options, nats.ReconnectWait(time.Duration(p.ReconnectDelay)*time.Second))
	}

	options = append(options, nats.MaxReconnects(p.MaxReconnects))

	options = append(options, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			clog.Warnf("disconnect error. [error = %v]", err)
		}
	}))

	options = append(options, nats.ReconnectHandler(func(nc *nats.Conn) {
		clog.Warnf("reconnected [%s]", nc.ConnectedUrl())
	}))

	options = append(options, nats.ClosedHandler(func(nc *nats.Conn) {
		clog.Infof("exiting... %s", p.Address)
		if nc.LastError() != nil {
			clog.Infof("error = %v", nc.LastError())
		}
	}))

	options = append(options, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		clog.Warnf("isConnect = %v. %s on connection for subscription on %q",
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
