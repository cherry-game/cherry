package cherryMessage

// Message types
const (
	TypeRequest  Type = 0x00 //客户端请求 ----000-
	TypeNotify   Type = 0x01 //客户端通知 ----001-
	TypeResponse Type = 0x02 //服务器返回 ----010-
	TypePush     Type = 0x03 //服务器推送 ----011-
)

//一些掩码定义用来操作flag(1byte)
const (
	MsgRouteCompressMask = 0x01 // 启用路由压缩 00000001
	MsgHeadLength        = 0x02 // 消息头的长度 00000010
	MsgTypeMask          = 0x07 // 获取消息类型 00000111
)

// Type represents the type of message, which could be TypeRequest/TypeNotify/TypeResponse/TypePush
type Type byte

var types = map[Type]string{
	TypeRequest:  "REQUEST",
	TypeNotify:   "Notify",
	TypeResponse: "Response",
	TypePush:     "Push",
}

func (t *Type) String() string {
	return types[*t]
}

func routable(t Type) bool {
	return t == TypeRequest || t == TypeNotify || t == TypePush
}

func invalidType(t Type) bool {
	return t < TypeRequest || t > TypePush
}
