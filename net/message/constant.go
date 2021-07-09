package cherryMessage

//一些掩码定义用来操作flag(1byte)
const (
	MsgHeadLength = 0x02 // 消息头的长度 00000010
	gzipMask      = 0x10 // data compressed gzip mark

	RouteCompressMask = 0x01 // 启用路由压缩 00000001
	ErrorMask         = 0x20 // 响应错误标识 00100000
	TypeMask          = 0x07 // 获取消息类型 00000111
)

var (
	dataCompression = false // encode message is compression
)

func SetDataCompression(compression bool) {
	dataCompression = compression
}
