package cherryInterfaces

import "net"

type (
	UID        = int64
	SID        = int64
	FrontendId = string

	SessionListener func(session ISession)
	MessageListener func(bytes []byte) error

	ISession interface {
		SID() SID
		UID() UID
		FrontendId() FrontendId
		SetStatus(status int)
		Status() int

		Data() map[string]interface{}
		Remove(key string)
		Set(key string, value interface{})
		Contains(key string) bool

		Conn() net.Conn
		OnClose(listener SessionListener)
		OnError(listener SessionListener)
		OnMessage(listener MessageListener)
		Send(msg []byte) error
		SendBatch(batchMsg ...[]byte)

		Closed()
	}

	INetworkEntity interface {
		Push(route string, data interface{}) error
		RPC(route string, data interface{}) error
		LastMid() uint64
		Response(data interface{}) error
		ResponseMid(mid uint64, data interface{}) error
		Close() error
		RemoteAddr() net.Addr
	}
)
