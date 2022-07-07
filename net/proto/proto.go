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
	req := requestPool.Get().(*Request)
	return req
}

func (x *Request) Recycle() {
	x.Reset()
	requestPool.Put(x)
}
