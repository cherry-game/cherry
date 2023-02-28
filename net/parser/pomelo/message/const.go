package pomeloMessage

// Message types
const (
	Request  Type = 0x00 // ----000-
	Notify   Type = 0x01 // ----001-
	Response Type = 0x02 // ----010-
	Push     Type = 0x03 // ----011-
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

var typeMap = map[Type]string{
	Request:  "Request",
	Notify:   "Notify",
	Response: "Response",
	Push:     "Push",
}

func (t *Type) String() string {
	return typeMap[*t]
}

func Routable(t Type) bool {
	return t == Request || t == Notify || t == Push
}

func InvalidType(t Type) bool {
	return t < Request || t > Push
}

// 掩码定义用来操作flag(1byte)
const (
	RouteCompressMask = 0x01 // 启用路由压缩 00000001
	MsgHeadLength     = 0x02 // 消息头的长度 00000010
	TypeMask          = 0x07 // 获取消息类型 00000111
	GZIPMask          = 0x10 // data compressed gzip mark
	ErrorMask         = 0x20 // 响应错误标识 00100000
)

var (
	dataCompression = false // encode message is compression
)

func IsDataCompression() bool {
	return dataCompression
}

func SetDataCompression(compression bool) {
	dataCompression = compression
}
