package pomeloClient

import (
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
)

type (
	options struct {
		serializer     cfacade.ISerializer // protocol serializer
		heartBeat      int                 // second
		requestTimeout time.Duration       // Send request timeout
		handshake      string              // handshake content
		isErrorBreak   bool                // an error occurs,is it break
	}

// Option is a functional option for configuring a Client.
	Option func(options *options)

// HandshakeSys is the system data sent by the server during handshake.
	HandshakeSys struct {
		Dict       map[string]uint16 `json:"dict"`
		Heartbeat  int               `json:"heartbeat"`
		Serializer string            `json:"serializer"`
	}

// HandshakeData is the complete handshake response from the server.
	// HandshakeData struct
	HandshakeData struct {
		Code int          `json:"code"`
		Sys  HandshakeSys `json:"sys"`
	}
)

// Serializer returns the configured serializer.
func (p *options) Serializer() cfacade.ISerializer {
	return p.serializer
}

// WithSerializer sets the protocol serializer for the client.
func WithSerializer(serializer cfacade.ISerializer) Option {
	return func(options *options) {
		options.serializer = serializer
	}
}

// WithHeartbeat sets the heartbeat interval in seconds.
func WithHeartbeat(heartBeat int) Option {
	return func(options *options) {
		options.heartBeat = heartBeat
	}
}

// WithRequestTimeout sets the timeout for request-response calls.
func WithRequestTimeout(requestTimeout time.Duration) Option {
	return func(options *options) {
		options.requestTimeout = requestTimeout
	}
}

// WithHandshake sets the handshake payload sent to the server.
func WithHandshake(handshake string) Option {
	return func(options *options) {
		options.handshake = handshake
	}
}

// WithErrorBreak sets whether the client disconnects when an action returns an error.
func WithErrorBreak(isBreak bool) Option {
	return func(options *options) {
		options.isErrorBreak = isBreak
	}
}
