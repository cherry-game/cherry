package cherryInterfaces

import "net"

type (
	UID             = int64                  // 用户唯一id
	SID             = int64                  // session唯一id
	FrontendId      = string                 // 前端结点id
	SessionListener func(session ISession)   // Session监听函数
	MessageListener func(bytes []byte) error // 消息监听函数

	// ISession 用户会话接口
	ISession interface {
		SID() SID                           // 用户唯一id
		UID() UID                           // session唯一id
		FrontendId() FrontendId             // 前端结点id
		SetStatus(status int)               // 设置session状态
		Status() int                        // 获取状态
		Data() map[string]interface{}       // session的k,v数据
		Remove(key string)                  // 移除数据
		Set(key string, value interface{})  // 设置数据
		Contains(key string) bool           // Data数据中是否包含该key
		Conn() net.Conn                     // 底层连接对象
		OnClose(listener SessionListener)   // 设置关闭时的监听函数
		OnError(listener SessionListener)   // 设置出错时的监听函数
		OnMessage(listener MessageListener) // 设置接收消息的监听函数
		Send(msg []byte) error              // 发送消息
		SendBatch(batchMsg ...[]byte)       // 批量发送消息
		Closed()                            // 关闭session
	}

	// INetworkEntity 网络处理接口
	INetworkEntity interface {
		Push(route string, data interface{}) error      // 推送消息对客户端
		RPC(route string, data interface{}) error       // 调用rpc
		LastMid() uint64                                // 获取消息序号
		Response(data interface{}) error                // 回复消息到客户端
		ResponseMid(mid uint64, data interface{}) error // 回复消息到客户端(带mid序号)
		Close() error                                   // 关闭接口
		RemoteAddr() net.Addr                           // 连接者的地址信息
	}
)
