package cherryFacade

import "net"

type (
	SID        = int64  // session唯一id
	UID        = string // 用户唯一id user unique id
	FrontendId = string // 前端节点id

	// INetwork 网络处理接口
	INetwork interface {
		SendRaw(bytes []byte) error
		Push(route string, v interface{}) error      // 推送消息对客户端
		RPC(route string, v interface{}) error       // 调用rpc
		ResponseMid(mid uint64, v interface{}) error // 回复消息到客户端
		Close()                                      // 关闭接口
		RemoteAddr() net.Addr                        // 连接者的地址信息
	}
)
