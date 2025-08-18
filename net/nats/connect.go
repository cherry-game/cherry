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
		*nats.Conn                    //
		options                       //
		id         int                //
		seq        uint64             //
		waiters    sync.Map           // map[string]chan *nats.Msg
		sub        *nats.Subscription //
		reply      string             // request reply subject
		replySub   *nats.Subscription //
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

func (p *Connect) Close() {
	if p.IsConnected() {
		p.sub.Unsubscribe()
		p.replySub.Unsubscribe()
		p.Conn.Close()
	}
}

func (p *Connect) statistics() {
	for {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			if p.sub != nil {
				if dropped, err := p.sub.Dropped(); err != nil {
					clog.Errorf("Dropped messages. [subject = %s, dropped = %d, err = %v]",
						p.sub.Subject,
						dropped,
						err,
					)
				}
			}

			if p.replySub != nil {
				if dropped, err := p.replySub.Dropped(); err != nil {
					clog.Errorf("Dropped messages. [subject = %s, dropped = %d, err = %v]",
						p.sub.Subject,
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

func (p *Connect) initReplySubscribe() {
	sub, err := p.Subscribe(p.reply, func(msg *nats.Msg) {
		reqID := msg.Header.Get(REQ_ID)
		if reqID == "" {
			clog.Infof("header = %v, subject = %v", msg.Header, msg.Subject)
			return
		}

		if chMsg, ok := p.waiters.LoadAndDelete(reqID); ok {
			ch := chMsg.(chan *nats.Msg)
			select {
			case ch <- msg:
			default:
			}
			close(ch)
		}
	})

	if err != nil {
		clog.Warnf(" err = %v", err)
		return
	}

	p.replySub = sub
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

	// get msg from pool
	msg := GetMsg()
	msg.Subject = subject
	msg.Reply = p.reply
	msg.Header.Set(REQ_ID, reqID)
	msg.Data = data

	err := p.PublishMsg(msg)

	// release msg
	ReleaseMsg(msg)

	if err != nil {
		p.waiters.Delete(reqID)
		close(ch)
		return nil, err
	}

	select {
	case resp, ok := <-ch:
		if !ok || resp == nil {
			return nil, cerror.ClusterRequestTimeout
		}
		return resp.Data, nil
	case <-time.After(timeout):
		p.waiters.Delete(reqID)
		clog.Warnf("id = %d, reqID = %s", p.id, reqID)
		close(ch)
		return nil, cerror.ClusterRequestTimeout
	}
}

func (p *Connect) QueueSubscribe(subject, queue string, cb nats.MsgHandler) error {
	sub, err := p.Conn.QueueSubscribe(subject, queue, cb)
	if err != nil {
		return err
	}

	p.sub = sub
	return nil
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

func WithIsStats(isStats bool) OptionFunc {
	return func(opts *options) {
		opts.isStats = isStats
	}
}
