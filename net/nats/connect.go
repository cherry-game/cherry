package cherryNats

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
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
		mu        sync.Mutex           // protect local mutable state: subs / stopStats
		id        int                  // connect id
		seq       uint64               // seq value
		waiters   sync.Map             // map[string]chan *nats.Msg
		subs      []*nats.Subscription // subscription list
		reply     string               // request reply subject
		stopStats chan struct{}        // notify statistics goroutine to exit
	}

	options struct {
		address       string        // NATS server address.
		maxReconnects int           // Maximum reconnect attempts handled by nats.Conn.
		user          string        // Optional NATS auth username.
		password      string        // Optional NATS auth password.
		isStats       bool          // Whether to start the statistics goroutine.
		statsInterval time.Duration // Statistics reporting interval. Defaults to 30 seconds when unset.
	}
	OptionFunc func(o *options)
)

func NewConnect(id int, replySubject string, opts ...OptionFunc) *Connect {
	conn := &Connect{
		id:    id,
		reply: fmt.Sprintf("%s.%d", replySubject, id),
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

	var conn *nats.Conn
	for {
		var err error
		conn, err = nats.Connect(p.address, p.natsOptions()...)
		if err != nil {
			clog.Warnf("[id = %d] Nats connect fail! retrying in 3 seconds. err = %s", p.id, err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	p.mu.Lock()
	if p.Conn != nil {
		p.mu.Unlock()
		conn.Close()
		return
	}
	p.Conn = conn
	if p.isStats && p.stopStats == nil {
		p.stopStats = make(chan struct{})
	}
	p.mu.Unlock()

	p.initReplySubscribe()

	if p.isStats {
		go p.statistics()
	}

	clog.Infof("[id = %d] Nats connected! [reply = %s]", p.id, p.reply)
}

func (p *Connect) Subs() []*nats.Subscription {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]*nats.Subscription(nil), p.subs...)
}

func (p *Connect) Close() {
	p.mu.Lock()
	if p.Conn == nil {
		p.mu.Unlock()
		return
	}

	stopStats := p.stopStats
	p.stopStats = nil
	subs := append([]*nats.Subscription(nil), p.subs...)
	p.subs = nil
	conn := p.Conn
	p.mu.Unlock()

	if stopStats != nil {
		close(stopStats)
	}

	for _, sub := range subs {
		sub.Unsubscribe()
	}

	conn.Close()
	p.clearWaiters()

	clog.Infof("[id = %d] Nats closed", p.id)
}

func (p *Connect) statistics() {
	p.mu.Lock()
	stopStats := p.stopStats
	p.mu.Unlock()
	if stopStats == nil {
		return
	}

	ticker := time.NewTicker(p.StatsInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			subs := p.Subs()

			for _, sub := range subs {
				if dropped, err := sub.Dropped(); err != nil {
					clog.Errorf("Dropped messages. [subject = %s, dropped = %d, err = %v]",
						sub.Subject,
						dropped,
						err,
					)
				}
			}

			stats := p.Stats()
			clog.Debugf("[Statistics] InMsgs = %d, OutMsgs = %d, InBytes = %d, OutBytes = %d, Reconnects = %d",
				stats.InMsgs,
				stats.OutMsgs,
				stats.InBytes,
				stats.OutBytes,
				stats.Reconnects,
			)
		case <-stopStats:
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
		ch := value.(chan *nats.Msg)
		p.waiters.Delete(key)
		close(ch)
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
		clog.Warnf(" err = %v", err)
		return
	}
}

func (p *Connect) Request(subject string, data []byte, tod ...time.Duration) ([]byte, error) {
	conn := p.Conn
	if conn == nil {
		return nil, fmt.Errorf("nats connection is nil")
	}

	timeout := GetTimeout(tod...)
	natsMsg, err := conn.Request(subject, data, timeout)
	if err != nil {
		return nil, err
	}

	return natsMsg.Data, nil
}

func (p *Connect) RequestSync(subject string, data []byte, tod ...time.Duration) ([]byte, error) {
	conn := p.Conn
	if conn == nil {
		return nil, fmt.Errorf("nats connection is nil")
	}

	timeout := GetTimeout(tod...)

	reqID := strconv.FormatUint(atomic.AddUint64(&p.seq, 1), 10)
	ch := make(chan *nats.Msg, 1)
	p.waiters.Store(reqID, ch)

	natsMsg := GetNatsMsg()
	natsMsg.Subject = subject
	natsMsg.Reply = p.reply
	natsMsg.Header.Set(REQ_ID, reqID)
	natsMsg.Data = data

	err := conn.PublishMsg(natsMsg.Msg)
	natsMsg.Release()

	if err != nil {
		p.waiters.Delete(reqID)
		close(ch)
		return nil, err
	}

	timer := acquireTimer(timeout)
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

func (p *Connect) Subscribe(subject string, cb nats.MsgHandler) error {
	conn := p.Conn
	if conn == nil {
		return fmt.Errorf("nats connection is nil")
	}

	sub, err := conn.Subscribe(subject, cb)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.subs = append(p.subs, sub)
	p.mu.Unlock()

	return nil
}

func (p *Connect) QueueSubscribe(subject, queue string, cb nats.MsgHandler) error {
	conn := p.Conn
	if conn == nil {
		return fmt.Errorf("nats connection is nil")
	}

	sub, err := conn.QueueSubscribe(subject, queue, cb)
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.subs = append(p.subs, sub)
	p.mu.Unlock()

	return nil
}

func (p *Connect) natsOptions() []nats.Option {
	var opts []nats.Option

	if reconnectDelay > 0 {
		opts = append(opts, nats.ReconnectWait(reconnectDelay))
	}

	if p.options.maxReconnects > 0 {
		opts = append(opts, nats.MaxReconnects(p.options.maxReconnects))
	}

	opts = append(opts, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			clog.Warnf("[id = %d] Disconnect error. [error = %v]", p.id, err)
		}
		p.clearWaiters()
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

func (p *options) StatsInterval() time.Duration {
	if p.statsInterval <= 0 {
		return 30 * time.Second
	}

	return p.statsInterval
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
