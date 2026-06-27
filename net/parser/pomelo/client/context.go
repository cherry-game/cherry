package pomeloClient

import (
	"time"

	cmsg "github.com/cherry-game/cherry/net/parser/pomelo/message"
)

// RequestContext carries the timeout ticker and response channel for a pending request.
type (
	RequestContext struct {
		*time.Ticker
		Chan chan *cmsg.Message
	}
)

// NewRequestContext creates a RequestContext with a ticker for the given timeout duration.
func NewRequestContext(t time.Duration) RequestContext {
	return RequestContext{
		Ticker: time.NewTicker(t),
		Chan:   make(chan *cmsg.Message, 1),
	}
}

// Close closes the response channel and stops the ticker.
func (p *RequestContext) Close() {
	if p.Chan != nil {
		close(p.Chan)
	}

	p.Ticker.Stop()
}
