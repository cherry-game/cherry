package cherryClient

import (
	cmsg "github.com/cherry-game/cherry/net/message"
	"time"
)

type (
	RequestContext struct {
		*time.Ticker
		Chan chan *cmsg.Message
	}
)

func NewRequestContext(t time.Duration) RequestContext {
	return RequestContext{
		Ticker: time.NewTicker(t),
		Chan:   make(chan *cmsg.Message, 1),
	}
}

func (p *RequestContext) Close() {
	if p.Chan != nil {
		close(p.Chan)
	}

	p.Ticker.Stop()
}
