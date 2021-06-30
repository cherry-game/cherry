package cherryFacade

import "net"

type (
	SID        = int64  // session唯一id
	UID        = int64  // 用户唯一id user unique id
	FrontendId = string // 前端节点id

	// INetwork 网络处理接口
	INetwork interface {
		Push(route string, val interface{}) error   // 推送消息对客户端
		Response(mid uint64, val interface{}) error // 回复消息到客户端
		Kick(reason string) error                   // 踢下线
		SendRaw(bytes []byte) error                 // write raw data to client
		RPC(route string, val interface{}) error    // 调用rpc
		Close()                                     // 关闭接口
		RemoteAddr() net.Addr                       // 连接者的地址信息
	}
)
