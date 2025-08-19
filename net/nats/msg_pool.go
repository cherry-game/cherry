package cherryNats

import (
	"sync"

	"github.com/nats-io/nats.go"
)

var (
	_msgPool = &sync.Pool{
		New: func() interface{} {
			return &nats.Msg{}
		},
	}
)

func GetMsg() *nats.Msg {
	value := _msgPool.Get()
	msg := value.(*nats.Msg)
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}

	return msg
}

func ReleaseMsg(msg *nats.Msg) {
	msg.Header = nil
	msg.Data = nil
	_msgPool.Put(msg)
}
