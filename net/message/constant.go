package cherryMessage

import (
	"github.com/cherry-game/cherry/extend/utils"
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

// Message types
const (
	Request  Type = 0x00 //客户端请求 ----000-
	Notify   Type = 0x01 //客户端通知 ----001-
	Response Type = 0x02 //服务器返回 ----010-
	Push     Type = 0x03 //服务器推送 ----011-
)

//一些掩码定义用来操作二进制
const (
	errorMask            = 0x20
	gzipMask             = 0x10
	msgRouteCompressMask = 0x01 //启用路由压缩
	msgTypeMask          = 0x07 //获取消息类型   00000111
	msgRouteLengthMask   = 0xFF //消息路由的长度 11111111
	msgHeadLength        = 0x02 //消息头的长度   00000010
)

var types = map[Type]string{
	Request:  "Request",
	Notify:   "Notify",
	Response: "Response",
	Push:     "Push",
}

var (
	routes = make(map[string]uint16) // route map to code  路由信息映射为uint16
	codes  = make(map[uint16]string) // code map to route  uint16映射为路由信息
)

// Errors that could be occurred in message codec
var (
	ErrWrongMessageType  = cherryUtils.Error("wrong message type")
	ErrInvalidMessage    = cherryUtils.Error("invalid message")
	ErrRouteInfoNotFound = cherryUtils.Error("route info not found in dictionary")
)
