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
	req.Reset()
	return req
}

func (x *Request) Recycle() {
	requestPool.Put(x)
}
