package cherryMessage

// Message types
const (
	Request  Type = 0x00 // ----000-
	Notify   Type = 0x01 // ----001-
	Response Type = 0x02 // ----010-
	Push     Type = 0x03 // ----011-
)

//一些掩码定义用来操作flag(1byte)
const (
	MsgRouteCompressMask = 0x01 // 启用路由压缩 00000001
	MsgHeadLength        = 0x02 // 消息头的长度 00000010
	MsgTypeMask          = 0x07 // 获取消息类型 00000111
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

var types = map[Type]string{
	Request:  "Request",
	Notify:   "Notify",
	Response: "Response",
	Push:     "Push",
}

func (t *Type) String() string {
	return types[*t]
}

func routable(t Type) bool {
	return t == Request || t == Notify || t == Push
}

func invalidType(t Type) bool {
	return t < Request || t > Push
}
