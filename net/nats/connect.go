package cherryNats

import (
	"fmt"
	"sync"
	"time"

	cerror "github.com/cherry-game/cherry/error"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/nats-io/nats.go"
)

const (
	REQ_ID = "reqID"
)

type (
	Connect struct {
		options
		*nats.Conn
		name      string               // connect name, passed in by the caller
		waiters   sync.Map             // map[string]chan *nats.Msg
		subs      []*nats.Subscription // subscription list
		subsMutex sync.RWMutex         // subscription mutex
		reply     string               // request reply subject, set by SubscribeReply
		stopStats chan struct{}        // notify statistics goroutine to exit
		closeOnce sync.Once            // ensure Close only executes once
	}

	options struct {
		address        string        // NATS server address.
		reconnectDelay time.Duration // Reconnect backoff interval. Defaults to 1 second when unset.
		requestTimeout time.Duration // Default request timeout. Defaults to 2 seconds when unset.
		maxReconnects  int           // Maximum reconnect attempts handled by nats.Conn.
		user           string        // Optional NATS auth username.
		password       string        // Optional NATS auth password.
		isStats        bool          // Whether to start the statistics goroutine.
		statsInterval  time.Duration // Statistics reporting interval. Defaults to 30 seconds when unset.
	}
	OptionFunc func(o *options)
)

func NewConnect(name string, opts ...OptionFunc) *Connect {
	conn := &Connect{
		name:      name,
		stopStats: make(chan struct{}),
	}

	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&conn.options)
		}
	}

	return conn
}

func (p *Connect) Connect() {
	if p.Conn != nil {
		return
	}

	var (
		natsOpts = p.natsOptions()
		conn     *nats.Conn
		err      error
	)

	for {
		conn, err = nats.Connect(p.address, natsOpts...)
		if err != nil {
			clog.Warnf("[name = %s] Nats connect fail! retrying in 3 seconds. err = %s", p.name, err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	p.Conn = conn

	if p.isStats {
		go p.statistics()
	}

	clog.Infof("[name = %s] Nats connected!", p.name)
}

func (p *Connect) Close() {
	p.closeOnce.Do(func() {
		if p.Conn == nil {
			return
		}

		p.subsMutex.Lock()
		for _, sub := range p.subs {
			sub.Unsubscribe()
		}
		p.subsMutex.Unlock()

		close(p.stopStats)

		p.Conn.Close()
		p.clearWaiters()

		clog.Infof("[name = %s] Nats closed", p.name)
	})
}

func (p *Connect) statistics() {
	ticker := time.NewTicker(p.StatsInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.subsMutex.RLock()
			for _, sub := range p.subs {
				if dropped, err := sub.Dropped(); err != nil {
					clog.Warnf("[%s] Dropped messages. [subject = %s, dropped = %d, err = %v]",
						p.name,
						sub.Subject,
						dropped,
						err,
					)
				}
			}
			p.subsMutex.RUnlock()

			stats := p.Stats()
			clog.Debugf("[%s] Statistics MaxPayload = %d, InMsgs = %d, OutMsgs = %d, InBytes = %d, OutBytes = %d, Reconnects = %d",
				p.name,
				p.MaxPayload(),
				stats.InMsgs,
				stats.OutMsgs,
				stats.InBytes,
				stats.OutBytes,
				stats.Reconnects,
			)
		case <-p.stopStats:
			clog.Infof("[%s] Statistics goroutine stopped", p.name)
			return
		}
	}
}

func (p *Connect) GetName() string {
	return p.name
}

func (p *Connect) clearWaiters() {
	p.waiters.Range(func(key, value any) bool {
		if v, ok := p.waiters.LoadAndDelete(key); ok {
			ch := v.(chan *nats.Msg)
			close(ch)
		}
		return true
	})
}

// SubscribeReply sets this connect's reply subject and subscribes to it, so
// that responses to RequestSync sent on this connect can be received. Call it
// after Connect() on connects that will issue requests; the subject base should
// be unique per node. Connects that only publish or subscribe do not need it.
func (p *Connect) SubscribeReply(replySubject string) error {
	if p.Conn == nil {
		return cerror.Error("nats connection is nil")
	}

	p.reply = fmt.Sprintf("%s.%s", replySubject, p.name)
	return p.Subscribe(p.reply, func(msg *nats.Msg) {
		reqID := msg.Header.Get(REQ_ID)
		if reqID == "" {
			clog.Infof("header = %v, subject = %v", msg.Header, msg.Subject)
			return
		}

		// LoadAndDelete takes ownership of the waiting channel.
		if chMsg, ok := p.waiters.LoadAndDelete(reqID); ok {
			ch := chMsg.(chan *nats.Msg)
			ch <- msg
			close(ch)
		}
	})
}

func (p *Connect) Request(subject string, data []byte, tod ...time.Duration) ([]byte, error) {
	if p.Conn == nil {
		return nil, fmt.Errorf("nats connection is nil")
	}

	timeout := p.Timeout(tod...)
	natsMsg, err := p.Conn.Request(subject, data, timeout)
	if err != nil {
		return nil, err
	}

	return natsMsg.Data, nil
}

func (p *Connect) RequestSync(reqID, subject string, data []byte, tod ...time.Duration) ([]byte, error) {
	if p.Conn == nil {
		return nil, fmt.Errorf("nats connection is nil")
	}
	if p.reply == "" {
		return nil, cerror.Error("reply subject is empty, call SubscribeReply first")
	}

	ch := make(chan *nats.Msg, 1)
	p.waiters.Store(reqID, ch)

	natsMsg := GetNatsMsg()
	natsMsg.Subject = subject
	natsMsg.Reply = p.reply
	natsMsg.Header.Set(REQ_ID, reqID)
	natsMsg.Data = data

	err := p.Conn.PublishMsg(natsMsg)
	ReleaseNatsMsg(natsMsg)
	if err != nil {
		if _, existed := p.waiters.LoadAndDelete(reqID); existed {
			close(ch)
		}
		return nil, err
	}

	timer := acquireTimer(p.Timeout(tod...))
	defer releaseTimer(timer)

	select {
	case resp, ok := <-ch:
		if !ok || resp == nil {
			return nil, cerror.ClusterRequestTimeout
		}
		return resp.Data, nil
	case <-timer.C:
		if _, existed := p.waiters.LoadAndDelete(reqID); existed {
			close(ch)
		}
		clog.Warnf("[RequestSync] timeout. name = %s, reqID = %s", p.name, reqID)
		return nil, cerror.ClusterRequestTimeout
	}
}

func (p *Connect) RequestReply(reqID, reply string, data []byte) error {
	if p.Conn == nil {
		return fmt.Errorf("nats connection is nil")
	}

	natsMsg := GetNatsMsg()
	natsMsg.Subject = reply
	natsMsg.Header.Set(REQ_ID, reqID)
	natsMsg.Data = data

	err := p.Conn.PublishMsg(natsMsg)
	ReleaseNatsMsg(natsMsg)
	return err
}

func (p *Connect) Subscribe(subject string, cb nats.MsgHandler) error {
	if p.Conn == nil {
		return cerror.Errorf("nats connection is nil. subject = %s", subject)
	}

	sub, err := p.Conn.Subscribe(subject, cb)
	if err != nil {
		return cerror.Errorf("Subscribe error. subject = %s, error = %v", subject, err)
	}

	p.subsMutex.Lock()
	p.subs = append(p.subs, sub)
	p.subsMutex.Unlock()

	return nil
}

func (p *Connect) QueueSubscribe(subject, queue string, cb nats.MsgHandler) error {
	if p.Conn == nil {
		return cerror.Errorf("nats connection is nil. subject = %s,queue = %s", subject, queue)
	}

	sub, err := p.Conn.QueueSubscribe(subject, queue, cb)
	if err != nil {
		return cerror.Errorf("QueueSubscribe error. subject = %s,queue = %s, error = %v", subject, queue, err)
	}

	p.subsMutex.Lock()
	p.subs = append(p.subs, sub)
	p.subsMutex.Unlock()

	return nil
}

func (p *Connect) natsOptions() []nats.Option {
	var opts []nats.Option

	if reconnectDelay := p.ReconnectDelay(); reconnectDelay > 0 {
		opts = append(opts, nats.ReconnectWait(reconnectDelay))
	}

	if p.options.maxReconnects > 0 {
		opts = append(opts, nats.MaxReconnects(p.options.maxReconnects))
	}

	opts = append(opts, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			clog.Warnf("[name = %s] Disconnect error. [error = %v]", p.name, err)
		}
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		clog.Warnf("[name = %s] Reconnected [%s]", p.name, nc.ConnectedUrl())
	}))

	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		if nc.LastError() != nil {
			clog.Infof("[name = %s] error = %v", p.name, nc.LastError())
		}
		p.clearWaiters()
	}))

	opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		clog.Warnf("[name = %s] IsConnect = %v. %s on connection for subscription on %q",
			p.name,
			nc.IsConnected(),
			err.Error(),
			sub.Subject,
		)
	}))

	if p.options.user != "" {
		opts = append(opts, nats.UserInfo(p.options.user, p.options.password))
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
	if p.reconnectDelay <= 0 {
		return 1 * time.Second
	}

	return p.reconnectDelay
}

func (p *options) RequestTimeout() time.Duration {
	if p.requestTimeout <= 0 {
		return 2 * time.Second
	}

	return p.requestTimeout
}

func (p *options) StatsInterval() time.Duration {
	if p.statsInterval <= 0 {
		return 60 * time.Second
	}

	return p.statsInterval
}

func (p *Connect) ReconnectDelay() time.Duration {
	return p.options.ReconnectDelay()
}

func (p *Connect) Timeout(tod ...time.Duration) time.Duration {
	if len(tod) > 0 {
		return tod[0]
	}

	return p.options.RequestTimeout()
}

func WithAddress(address string) OptionFunc {
	return func(opts *options) {
		opts.address = address
	}
}

func WithReconnectDelay(delay time.Duration) OptionFunc {
	return func(opts *options) {
		opts.reconnectDelay = delay
	}
}

func WithRequestTimeout(timeout time.Duration) OptionFunc {
	return func(opts *options) {
		opts.requestTimeout = timeout
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

func WithIsStats(isStats bool) OptionFunc {
	return func(opts *options) {
		opts.isStats = isStats
	}
}

func WithStatsInterval(seconds int) OptionFunc {
	return func(opts *options) {
		opts.statsInterval = time.Duration(seconds) * time.Second
	}
}
