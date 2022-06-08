package cherryMessage

// 掩码定义用来操作flag(1byte)
const (
	routeCompressMask = 0x01 // 启用路由压缩 00000001
	msgHeadLength     = 0x02 // 消息头的长度 00000010
	typeMask          = 0x07 // 获取消息类型 00000111
	gzipMask          = 0x10 // data compressed gzip mark
	errorMask         = 0x20 // 响应错误标识 00100000
)

var (
	dataCompression = false // encode message is compression
)

func SetDataCompression(compression bool) {
	dataCompression = compression
}
