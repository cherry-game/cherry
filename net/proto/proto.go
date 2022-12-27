package cherryProto

import (
	"sync"
)

var (
	requestPool = &sync.Pool{
		New: func() interface{} {
			return new(Request)
		},
	}
)

func GetRequest() *Request {
	request := requestPool.Get().(*Request)
	return request
}

func (m *Request) Recycle() {
	m.Reset()
	requestPool.Put(m)
}
