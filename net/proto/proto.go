package cherryProto

import (
	cherryString "github.com/cherry-game/cherry/extend/string"
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
	req := requestPool.Get().(*Request)
	if req.Setting == nil {
		req.Setting = make(map[string]string)
	}

	tt := cherryTime.Now().ToTimestampWithMillisecond()
	req.Setting[StartTimeKey] = cherryString.ToString(tt)

	return req
}

func (x *Request) Recycle() {
	x.Reset()
	requestPool.Put(x)
}
