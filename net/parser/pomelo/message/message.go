package pomeloMessage

import (
	"encoding/binary"
	"fmt"

	cerr "github.com/cherry-game/cherry/error"
	ccompress "github.com/cherry-game/cherry/extend/compress"
)

var (
	nilMessage = Message{}
)

// Message represents a unmarshaled message or a message which to be marshaled
// message协议的主要作用是封装消息头，包括route和消息类型两部分，
// 不同的消息类型有着不同的消息头，在消息头里面可能要打入message id(即requestID)和route信息。
// 由于可能会有route压缩，而且对于服务端push的消息，message id为空，对于客户端请求的响应，route为空
// 消息头分为三部分，flag，message id，route。
// 如下图所示：
// flag(1byte) + message id(0~5byte) + route(0~256bytes)
// flag位是必须的，占用一个byte，它决定了后面的消息类型和内容的格式;
// message id和route则是可选的。
// 其中message id采用varints 128变长编码方式，根据值的大小，长度在0～5byte之间。
// route则根据消息类型以及内容的大小，长度在0～255byte之间。
//
// flag占用message头的第一个byte，其内容如下
// preserved（4bits） + message type(3 bits) + route(1bit)
// 现在只用到了其中的4个bit，这四个bit包括两部分，占用3个bit的message type字段和占用1个bit的route标识，其中：
// message type用来标识消息类型,范围为0～7，
//
// 消息类型: 不同类型的消息，对应不同消息头，消息类型通过flag字段的第2-4位来确定，其对应关系以及相应的消息头如下图：
//
// 现在消息共有四类，request，notify，response，push，值的范围是0～3。
// 不同的消息类型有着不同的消息内容，下面会有详细分析。
// 最后一位的route表示route是否压缩，影响route字段的长度。 这两部分之间相互独立，互不影响。
// request   ----000-  <message id> <route>
// notify    ----001-  <route>
// response  ----010-  <message id>
// push      ----011-  <route>
//
// 路由压缩标志
// 上图是不同的flag标志对应的route字段的内容：
// flag的最后一位为1时，表示路由压缩，需要通过查询字典来获取route;
// flag最后一位为0是，后面route则由一个uInt8的byte，用来表示route的字节长度。
// 之后是通过utf8编码后的route字 符串，其长度就是前面一位byte的uInt8的值，因此route的长度最大支持256B。
type Message struct {
	Type            Type   // message type 4中消息类型
	ID              uint   // unique id, zero while notify mode 消息id（request response）
	Route           string // route for locating service 消息路由
	Data            []byte // payload  消息体的原始数据
	routeCompressed bool   // is route Compressed 是否启用路由压缩
	Error           bool   // response error
}

func New() Message {
	return Message{}
}

func (t *Message) String() string {
	return fmt.Sprintf(
		"Type: %s, ID: %d, Route: %s, RouteCompressed: %t, Data: %v, BodyLength: %d, Error:%v",
		t.Type.String(),
		t.ID,
		t.Route,
		t.routeCompressed,
		t.Data,
		len(t.Data),
		t.Error)
}

// Encode marshals message to binary format. Different message types is corresponding to
// different message header, message types is identified by 2-4 bit of flag field. The
// relationship between message types and message header is presented as follows:
// ------------------------------------------
// |   type   |  flag  |       other        |
// |----------|--------|--------------------|
// | request  |----000-|<message id>|<route>|
// | notify   |----001-|<route>             |
// | response |----010-|<message id>        |
// | push     |----011-|<route>             |
// ------------------------------------------
// The figure above indicates that the bit does not affect the type of message.
// See ref: https://github.com/lonnng/nano/blob/master/docs/communication_protocol.md
// See ref: https://github.com/NetEase/pomelo/wiki/%E5%8D%8F%E8%AE%AE%E6%A0%BC%E5%BC%8F
func Encode(m *Message) ([]byte, error) {
	if InvalidType(m.Type) {
		return nil, cerr.MessageWrongType
	}

	buf := make([]byte, 0)
	flag := byte(m.Type) << 1

	code, compressed := GetCode(m.Route)

	if compressed {
		flag |= RouteCompressMask
	}

	if m.Error {
		flag |= ErrorMask
	}

	buf = append(buf, flag)

	if m.Type == Request || m.Type == Response {
		n := m.ID
		// variant length encode
		for {
			b := byte(n % 128)
			n >>= 7
			if n != 0 {
				buf = append(buf, b+128)
			} else {
				buf = append(buf, b)
				break
			}
		}
	}

	if Routable(m.Type) {
		if compressed {
			buf = append(buf, byte((code>>8)&0xFF))
			buf = append(buf, byte(code&0xFF))
		} else {
			buf = append(buf, byte(len(m.Route)))
			buf = append(buf, []byte(m.Route)...)
		}
	}

	if IsDataCompression() {
		d, err := ccompress.DeflateData(m.Data)
		if err != nil {
			return nil, err
		}

		if len(d) < len(m.Data) {
			m.Data = d
			buf[0] |= GZIPMask
		}
	}

	buf = append(buf, m.Data...)
	return buf, nil
}

// Decode unmarshal the bytes slice to a message
// See ref: https://github.com/lonnng/nano/blob/master/docs/communication_protocol.md
func Decode(data []byte) (Message, error) {
	if len(data) < MsgHeadLength {
		return nilMessage, cerr.MessageInvalid
	}

	m := New()
	flag := data[0]
	offset := 1
	m.Type = Type((flag >> 1) & TypeMask)

	if InvalidType(m.Type) {
		return nilMessage, cerr.MessageWrongType
	}

	if m.Type == Request || m.Type == Response {
		id := uint(0)
		// little end byte order
		// WARNING: must can be stored in 64 bits integer
		// variant length encode
		for i := offset; i < len(data); i++ {
			b := data[i]
			id += uint(b&0x7F) << uint(7*(i-offset))
			if b < 128 {
				offset = i + 1
				break
			}
		}
		m.ID = id
	}

	if offset > len(data) {
		return nilMessage, cerr.MessageInvalid
	}

	m.Error = flag&ErrorMask == ErrorMask

	if Routable(m.Type) {
		if flag&RouteCompressMask == 1 {
			m.routeCompressed = true
			code := binary.BigEndian.Uint16(data[offset:(offset + 2)])
			route, found := GetRoute(code)
			if !found {
				return nilMessage, cerr.MessageRouteNotFound
			}
			m.Route = route
			offset += 2

		} else {
			m.routeCompressed = false
			rl := data[offset]
			offset++
			m.Route = string(data[offset:(offset + int(rl))])
			offset += int(rl)
		}
	}

	if offset > len(data) {
		return nilMessage, cerr.MessageInvalid
	}

	m.Data = data[offset:]

	var err error
	if flag&GZIPMask == GZIPMask {
		m.Data, err = ccompress.InflateData(m.Data)
		if err != nil {
			return nilMessage, err
		}
	}

	return m, nil
}
