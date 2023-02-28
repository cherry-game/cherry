package pomeloClient

import (
	cfacade "github.com/cherry-game/cherry/facade"
	"time"
)

type (
	options struct {
		serializer     cfacade.ISerializer // protocol serializer
		heartBeat      int                 // second
		requestTimeout time.Duration       // Send request timeout
		handshake      string              // handshake content
		isErrorBreak   bool                // an error occurs,is it break
	}

	Option func(options *options)

	// HandshakeSys struct
	HandshakeSys struct {
		Dict       map[string]uint16 `json:"dict"`
		Heartbeat  int               `json:"heartbeat"`
		Serializer string            `json:"serializer"`
	}

	// HandshakeData struct
	HandshakeData struct {
		Code int          `json:"code"`
		Sys  HandshakeSys `json:"sys"`
	}
)

func (p *options) Serializer() cfacade.ISerializer {
	return p.serializer
}

func WithSerializer(serializer cfacade.ISerializer) Option {
	return func(options *options) {
		options.serializer = serializer
	}
}

func WithHeartbeat(heartBeat int) Option {
	return func(options *options) {
		options.heartBeat = heartBeat
	}
}

func WithRequestTimeout(requestTimeout time.Duration) Option {
	return func(options *options) {
		options.requestTimeout = requestTimeout
	}
}

func WithHandshake(handshake string) Option {
	return func(options *options) {
		options.handshake = handshake
	}
}

func WithErrorBreak(isBreak bool) Option {
	return func(options *options) {
		options.isErrorBreak = isBreak
	}
}
