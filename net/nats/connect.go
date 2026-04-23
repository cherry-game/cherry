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
		id        int                  // connect id
		waiters   sync.Map             // map[string]chan *nats.Msg
		subs      []*nats.Subscription // subscription list
		subsMutex sync.RWMutex         // subscription mutex
		reply     string               // request reply subject
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

func NewConnect(id int, replySubject string, opts ...OptionFunc) *Connect {
	conn := &Connect{
		id:        id,
		reply:     fmt.Sprintf("%s.%d", replySubject, id),
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
			clog.Warnf("[id = %d] Nats connect fail! retrying in 3 seconds. err = %s", p.id, err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	p.Conn = conn

	p.initReplySubscribe()

	if p.isStats {
		go p.statistics()
	}

	clog.Infof("[id = %d] Nats connected! [reply = %s]", p.id, p.reply)
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

		clog.Infof("[id = %d] Nats closed", p.id)
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
					clog.Errorf("Dropped messages. [subject = %s, dropped = %d, err = %v]",
						sub.Subject,
						dropped,
						err,
					)
				}
			}
			p.subsMutex.RUnlock()

			stats := p.Stats()
			clog.Debugf("[Statistics] InMsgs = %d, OutMsgs = %d, InBytes = %d, OutBytes = %d, Reconnects = %d",
				stats.InMsgs,
				stats.OutMsgs,
				stats.InBytes,
				stats.OutBytes,
				stats.Reconnects,
			)
		case <-p.stopStats:
			clog.Infof("[id = %d] Statistics goroutine stopped", p.id)
			return
		}
	}
}

func (p *Connect) GetID() int {
	return p.id
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

func (p *Connect) initReplySubscribe() {
	err := p.Subscribe(p.reply, func(msg *nats.Msg) {
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

	if err != nil {
		clog.Warnf("[initReplySubscribe] error = %v", err)
	}

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
		clog.Warnf("[RequestSync] timeout. id = %d, reqID = %s", p.id, reqID)
		return nil, cerror.ClusterRequestTimeout
	}
}

func (p *Connect) ReplySync(reqID, reply string, data []byte) error {
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
			clog.Warnf("[id = %d] Disconnect error. [error = %v]", p.id, err)
		}
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		clog.Warnf("[id = %d] Reconnected [%s]", p.id, nc.ConnectedUrl())
	}))

	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		if nc.LastError() != nil {
			clog.Infof("[id = %d] error = %v", p.id, nc.LastError())
		}
		p.clearWaiters()
	}))

	opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
		clog.Warnf("[id = %d] IsConnect = %v. %s on connection for subscription on %q",
			p.id,
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
		return 30 * time.Second
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
