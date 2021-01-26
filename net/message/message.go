package cherryMessage

import (
	"fmt"
	"strings"
)

// message协议的主要作用是封装消息头，包括route和消息类型两部分，
// 不同的消息类型有着不同的消息头，在消息头里面可能要打入message id(即requestId)和route信息。
// 由于可能会有route压缩，而且对于服务端push的消息，message id为空，对于客户端请求的响应，route为空
// 消息头分为三部分，flag，message id，route。
//如下所示：
// flag(1byte) + message id(0~5byte) + route(0~256bytes)
// flag位是必须的，占用一个byte，它决定了后面的消息类型和内容的格式;
// message id和route则是可选的。
// 其中message id采用varints 128变长编码方式，根据值的大小，长度在0～5byte之间。
// route则根据消息类型以及内容的大小，长度在0～255byte之间。

// flag占用message头的第一个byte，其内容如下
// preserved（4bits） + message type(3 bits) + route(1bit)
// 现在只用到了其中的4个bit，这四个bit包括两部分，占用3个bit的message type字段和占用1个bit的route标识，其中：
// message type用来标识消息类型,范围为0～7，

// 消息类型: 不同类型的消息，对应不同消息头，消息类型通过flag字段的第2-4位来确定，其对应关系以及相应的消息头如下图：

// 现在消息共有四类，request，notify，response，push，值的范围是0～3。
// 不同的消息类型有着不同的消息内容，下面会有详细分析。
// 最后一位的route表示route是否压缩，影响route字段的长度。 这两部分之间相互独立，互不影响。
// request   ----000-  <message id> <route>
// notify    ----001-  <route>
// response  ----010-  <message id>
// push      ----011-  <route>

// 路由压缩标志
// 上图是不同的flag标志对应的route字段的内容：
// flag的最后一位为1时，后面跟的是一个uInt16表示的route字典编号，需要通过查询字典来获取route;
// flag最后一位为0是，后面route则由一个uInt8的byte，用来表示route的字节长度。
// 之后是通过utf8编码后的route字 符串，其长度就是前面一位byte的uInt8的值，因此route的长度最大支持256B。

// Message represents a unmarshaled message or a message which to be marshaled
type Message struct {
	Type       Type   // message type 4中消息类型
	ID         uint   // unique id, zero while notify mode 消息id（request response）
	Route      string // route for locating service 消息路由
	Data       []byte // payload  消息体的原始数据
	compressed bool   // is message compressed 是否启用路由压缩
	Err        bool   // is an error message
}

func New(err ...bool) *Message {
	m := &Message{}
	if len(err) > 0 {
		m.Err = err[0]
	}
	return m
}

func (t *Message) String() string {
	return fmt.Sprintf(
		"Type: %s, ID: %d, Route: %s, Compressed: %t, Error: %t, Settings: %v, BodyLength: %d",
		types[t.Type],
		t.ID,
		t.Route,
		t.compressed,
		t.Err,
		t.Data,
		len(t.Data))
}

func routable(t Type) bool {
	return t == Request || t == Notify || t == Push
}

func invalidType(t Type) bool {
	return t < Request || t > Push
}

//启用路由压缩
//对于服务端，server会扫描所有的Handler信息
//对于客户端，用户需要配置一个路由映射表
//通过这两种方式，pitaya会拿到所有的客户端和服务端的路由信息，然后将每一个路由信息都映射为一个小整数，
//在客户端与服务器建立连接的握手过程中，服务器会将 整个字典传给客户端，
//这样在以后的通信中，对于路由信息，将全部使用定义的小整数进行标记，大大地减少了额外信 息开销

// SetDictionary set routes map which be used to compress route.
func SetDictionary(dict map[string]uint16) error {
	if dict == nil {
		return nil
	}

	for route, code := range dict {
		r := strings.TrimSpace(route) //去掉开头结尾的空格

		// duplication check
		if _, ok := routes[r]; ok {
			return fmt.Errorf("duplicated route(route: %s, code: %d)", r, code)
		}

		if _, ok := codes[code]; ok {
			return fmt.Errorf("duplicated route(route: %s, code: %d)", r, code)
		}

		// update map, using last value when key duplicated
		routes[r] = code
		codes[code] = r
	}
	return nil
}

// GetDictionary gets the routes map which is used to compress route.
func GetDictionary() map[string]uint16 {
	return routes
}

func (t *Type) String() string {
	return types[*t]
}
