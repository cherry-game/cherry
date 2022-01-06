package cherryFacade

import (
	cherryProto "github.com/cherry-game/cherry/net/proto"
)

type (
	SID        = string // session unique id
	UID        = int64  // 用户唯一id user unique id
	FrontendId = string // 前端节点id

	// INetwork 网络处理接口
	INetwork interface {
		Push(route string, val interface{})                           // 推送消息对客户端
		Kick(reason interface{})                                      // 踢下线
		Response(mid uint, val interface{}, isError ...bool)          // 回复消息到客户端
		RPC(route string, val interface{}, rsp *cherryProto.Response) // 调用rpc
		SendRaw(bytes []byte)                                         // write raw data to client
		RemoteAddr() string                                           // 连接者的地址信息
		Close()                                                       // 关闭接口
	}
)
