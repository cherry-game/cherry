package cherryNats

import (
	"time"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/nats-io/nats.go"
)

type (
	Connect struct {
		*nats.Conn
		options
		running bool
		index   int
	}

	options struct {
		address       string
		maxReconnects int
		user          string
		password      string
	}
	OptionFunc func(o *options)
)

func New(opts ...OptionFunc) *Connect {
	conn := &Connect{}

	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&conn.options)
		}
	}

	return conn
}

func (p *Connect) Connect() {
	if p.running {
		return
	}

	for {
		conn, err := nats.Connect(p.address, p.natsOptions()...)
		if err != nil {
			clog.Warnf("[%d] nats connect fail! retrying in 3 seconds. err = %s", p.index, err)
			time.Sleep(3 * time.Second)
			continue
		}

		p.Conn = conn
		p.running = true
		clog.Infof("[%d] nats is connected! [address = %s]", p.index, p.address)
		break
	}
}

func (p *Connect) Close() {
	if p.running {
		p.running = false
		p.Conn.Close()
		clog.Infof("[%d] nats connect execute Close()", p.index)
	}
}

func (p *Connect) GetIndex() int {
	return p.index
}

func (p *Connect) Request(subj string, data []byte, timeout ...time.Duration) (*nats.Msg, error) {
	if len(timeout) > 0 && timeout[0] > 0 {
		return p.Conn.Request(subj, data, timeout[0])
	}

	return p.Conn.Request(subj, data, requestTimeout)
}

func (p *Connect) ChanExecute(subject string, msgChan chan *nats.Msg, process func(msg *nats.Msg)) {
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

	if reconnectDelay > 0 {
		opts = append(opts, nats.ReconnectWait(reconnectDelay))
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

func WithAddress(address string) OptionFunc {
	return func(opts *options) {
		opts.address = address
	}
}

func WithParams(maxReconnects int) OptionFunc {
	return func(opts *options) {
		opts.maxReconnects = maxReconnects
	}
}

func WithAuth(user, password string) OptionFunc {
	return func(opts *options) {
		opts.user = user
		opts.password = password
	}
}
