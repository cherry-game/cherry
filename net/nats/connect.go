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
		*nats.Conn                      //
		options                         //
		id         int                  //
		seq        uint64               //
		waiters    sync.Map             // map[string]chan *nats.Msg
		subs       []*nats.Subscription //
		reply      string               // request reply subject
	}

	options struct {
		address       string
		maxReconnects int
		user          string
		password      string
		isStats       bool
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

	for {
		conn, err := nats.Connect(p.address, p.natsOptions()...)
		if err != nil {
			clog.Warnf("[id = %d] Nats connect fail! retrying in 3 seconds. err = %s", p.id, err)
			time.Sleep(3 * time.Second)
			continue
		}
		p.Conn = conn
		p.initReplySubscribe()

		if p.isStats {
			go p.statistics()
		}

		break
	}
}

func (p *Connect) Subs() []*nats.Subscription {
	return p.subs
}

func (p *Connect) Close() {
	if p.IsConnected() {
		for _, sub := range p.subs {
			sub.Unsubscribe()
		}

		p.Conn.Close()

		// 连接关闭后清理残留的 waiters（防御性编程）
		p.clearWaiters()
	}
}

func (p *Connect) statistics() {
	for {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			for _, sub := range p.subs {
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

		// LoadAndDelete 获取 channel 所有权，如果不存在说明请求方已超时处理
		if chMsg, ok := p.waiters.LoadAndDelete(reqID); ok {
			ch := chMsg.(chan *nats.Msg)
			// 直接发送到 buffered channel，请求方已准备好等待
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
	timeout := GetTimeout(tod...)
	natsMsg, err := p.Conn.Request(subject, data, timeout)
	if err != nil {
		return nil, err
	}

	return natsMsg.Data, nil
}

func (p *Connect) RequestSync(subject string, data []byte, tod ...time.Duration) ([]byte, error) {
	timeout := GetTimeout(tod...)

	reqID := strconv.FormatUint(atomic.AddUint64(&p.seq, 1), 10)
	ch := make(chan *nats.Msg, 1)
	p.waiters.Store(reqID, ch)

	natsMsg := GetNatsMsg()
	natsMsg.Subject = subject
	natsMsg.Reply = p.reply
	natsMsg.Header.Set(REQ_ID, reqID)
	natsMsg.Data = data

	err := p.PublishMsg(natsMsg.Msg)
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
	sub, err := p.Conn.Subscribe(subject, cb)
	if err != nil {
		return err
	}

	if sub != nil {
		p.subs = append(p.subs, sub)
	}

	return nil
}

func (p *Connect) QueueSubscribe(subject, queue string, cb nats.MsgHandler) error {
	sub, err := p.Conn.QueueSubscribe(subject, queue, cb)
	if err != nil {
		return err
	}

	if sub != nil {
		p.subs = append(p.subs, sub)
	}

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
		// 连接断开时清理所有等待中的请求
		p.clearWaiters()
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		clog.Warnf("[id = %d] Reconnected [%s]", p.id, nc.ConnectedUrl())
	}))

	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		if nc.LastError() != nil {
			clog.Infof("[id = %d] error = %v", p.id, nc.LastError())
		}
		// 连接关闭时清理所有等待中的请求
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
