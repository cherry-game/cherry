package cherryNats

import (
	"sync"

	"github.com/nats-io/nats.go"
)

var (
	_msgPool = &sync.Pool{
		New: func() any {
			return &NatsMsg{Msg: &nats.Msg{}}
		},
	}
)

type NatsMsg struct {
	*nats.Msg
}

func GetNatsMsg() *NatsMsg {
	msg := _msgPool.Get().(*NatsMsg)
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}
	return msg
}

func (m *NatsMsg) Release() {
	for k := range m.Header {
		delete(m.Header, k)
	}
	m.Data = nil
	_msgPool.Put(m)
}
