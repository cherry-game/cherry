package cherryProto

import (
	cherryTime "github.com/cherry-game/cherry/extend/time"
	"sync"
)

var (
	StartTimeKey = "1"
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
	if request.Setting == nil {
		request.Setting = make(map[string]string)
	}
	request.Setting[StartTimeKey] = cherryTime.Now().ToMillisecondString()

	return request
}

func (m *Request) Recycle() {
	m.Reset()
	requestPool.Put(m)
}
