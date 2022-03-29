package cherryClient

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	"time"
)

var (
	defaultHandshakeBuffer = `
{
	"sys": {
		"platform": "mac",
		"libVersion": "0.3.5-release",
		"clientBuildNumber":"20",
		"clientVersion":"2.1"
	}
}
`
)

type (
	options struct {
		serializer     cherryFacade.ISerializer  // protocol serializer
		codec          cherryFacade.IPacketCodec // packet codec
		heartBeat      int                       // second
		requestTimeout time.Duration             // Send request timeout
		handshake      string                    // hand shake content
		isErrorBreak   bool                      // an error occurs,is it break
	}

	Option func(options *options)

	// HandshakeSys struct
	HandshakeSys struct {
		Dict      map[string]uint16 `json:"dict"`
		Heartbeat int               `json:"heartbeat"`
	}

	// HandshakeData struct
	HandshakeData struct {
		Code int          `json:"code"`
		Sys  HandshakeSys `json:"sys"`
	}
)

func (p *options) Serializer() cherryFacade.ISerializer {
	return p.serializer
}

func (p *options) Codec() cherryFacade.IPacketCodec {
	return p.codec
}

func WithSerializer(serializer cherryFacade.ISerializer) Option {
	return func(options *options) {
		options.serializer = serializer
	}
}

func WithPacketCodec(codec cherryFacade.IPacketCodec) Option {
	return func(options *options) {
		options.codec = codec
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
