package cherryContext

// context key

type propagateKey struct{}

// PropagateCtxKey is the context key where the content that will be
// propagated through rpc calls is set
var PropagateCtxKey = propagateKey{}

const (
	MessageIdKey       = "message_id"        // (客户端)消息id
	BuildPacketTimeKey = "build_packet_time" // 创建数据包时间
	InHandlerTimeKey   = "in_handler_time"   // 进入handler时间
)
