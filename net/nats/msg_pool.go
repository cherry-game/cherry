package cherryNats

import (
	"sync"

	"github.com/nats-io/nats.go"
)

var (
	_natsMsgPool = &sync.Pool{
		New: func() any {
			return &nats.Msg{}
		},
	}
)

func GetNatsMsg() *nats.Msg {
	msg := _natsMsgPool.Get().(*nats.Msg)
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}
	return msg
}

func ReleaseNatsMsg(natsMsg *nats.Msg) {
	natsMsg.Header = nil
	natsMsg.Subject = ""
	natsMsg.Reply = ""
	natsMsg.Data = nil
	_natsMsgPool.Put(natsMsg)
}
