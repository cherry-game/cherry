package cherryNats

import (
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"time"
)

var (
	_conn = &Conn{
		running: false,
	}
)

type (
	Conn struct {
		*nats.Conn
		options
		running bool
	}

	options struct {
		address        string
		maxReconnects  int
		reconnectDelay time.Duration
		requestTimeout time.Duration
		user           string
		password       string
	}
	OptionFunc func(o *options)
)

func New(opts ...OptionFunc) *Conn {
	if _conn.running {
		return _conn
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&_conn.options)
		}
	} else {
		_conn.loadFromConfig()
	}

	return _conn
}

func (p *Conn) loadFromConfig() {
	config := cprofile.GetConfig("cluster").GetConfig("nats")
	if config.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	p.address = config.GetString("address")
	p.maxReconnects = config.GetInt("max_reconnects")
	p.reconnectDelay = config.GetDuration("reconnect_delay", 1) * time.Second
	p.requestTimeout = config.GetDuration("request_timeout", 1) * time.Second
	p.user = config.GetString("user")
	p.password = config.GetString("password")

	if p.address == "" {
		panic("address is empty!")
	}
}

func (p *Conn) Connect() {
	if p.running {
		return
	}

	for {
		conn, err := nats.Connect(p.address, p.natsOptions()...)
		if err != nil {
			clog.Warnf("nats connect fail! retrying in 3 seconds. err = %s", err)
			time.Sleep(3 * time.Second)
			continue
		}

		p.Conn = conn
		p.running = true
		clog.Infof("nats is connected! [address = %s]", p.address)
		break
	}
}

func (p *Conn) Close() {
	if p.running {
		p.running = false
		p.Conn.Close()
		clog.Infof("nats connect execute Close()")
	}
}

func (p *Conn) Request(subj string, data []byte, timeout ...time.Duration) (*nats.Msg, error) {
	if len(timeout) > 0 && timeout[0] > 0 {
		return p.Conn.Request(subj, data, timeout[0])
	}

	return p.Conn.Request(subj, data, p.requestTimeout)
}

func (p *Conn) ChanExecute(subject string, msgChan chan *nats.Msg, process func(msg *nats.Msg)) {
	_, chanErr := p.ChanSubscribe(subject, msgChan)
	if chanErr != nil {
		clog.Error("subscribe fail. [subject = %s, err = %s]", subject, chanErr)
		return
	}

	for msg := range msgChan {
		process(msg)
	}
}

func (p *options) natsOptions() []nats.Option {
	var opts []nats.Option

	if p.reconnectDelay > 0 {
		opts = append(opts, nats.ReconnectWait(p.reconnectDelay))
	}

	if p.maxReconnects > 0 {
		opts = append(opts, nats.MaxReconnects(p.maxReconnects))
	}

	opts = append(opts, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			clog.Warnf("Disconnect error. [error = %v]", err)
		}
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		clog.Warnf("Reconnected [%s]", nc.ConnectedUrl())
	}))

	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		clog.Infof("Nats exiting... %s", p.address)
		if nc.LastError() != nil {
			clog.Infof("error = %v", nc.LastError())
		}
	}))

	opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		clog.Warnf("IsConnect = %v. %s on connection for subscription on %q",
			nc.IsConnected(),
			err.Error(),
			sub.Subject,
		)
	}))

	if p.user != "" {
		opts = append(opts, nats.UserInfo(p.user, p.password))
	}

	return opts
}

func (p *options) Address() string {
	return p.address
}

func (p *options) MaxReconnects() int {
	return p.maxReconnects
}

func (p *options) ReconnectDelay() time.Duration {
	return p.reconnectDelay
}

func (p *options) RequestTimeout() time.Duration {
	return p.requestTimeout
}

func WithAddress(address string) OptionFunc {
	return func(opts *options) {
		opts.address = address
	}
}

func WithParams(maxReconnects int, reconnectDelay int, requestTimeout int) OptionFunc {
	return func(opts *options) {
		opts.maxReconnects = maxReconnects
		opts.reconnectDelay = time.Duration(reconnectDelay) * time.Second
		opts.requestTimeout = time.Duration(requestTimeout) * time.Second
	}
}

func WithAuth(user, password string) OptionFunc {
	return func(opts *options) {
		opts.user = user
		opts.password = password
	}
}
