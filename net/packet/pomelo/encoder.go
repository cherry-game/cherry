package cherryPacketPomelo

import (
	"github.com/cherry-game/cherry/net/packet"
)

type Encoder struct {
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

// PacketEncode create a packet.Packet from  the raw bytes slice and then encode to network bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func (p *Encoder) Encode(typ byte, data []byte) ([]byte, error) {
	if typ < Handshake || typ > Kick {
		return nil, ErrWrongPomeloPacketType
	}

	if len(data) > MaxPacketSize {
		return nil, ErrPacketSizeExcced
	}

	pkg := &cherryPacket.Packet{
		Type:   typ,
		Length: len(data),
	}

	//生成一个切片长度=消息头长度+消息体长度 header+body = 4 + len(body)
	buf := make([]byte, pkg.Length+HeadLength)

	//第一个字节存放消息类型
	buf[0] = pkg.Type

	//2~4 字节 存放消息长度
	copy(buf[1:HeadLength], intToBytes(pkg.Length))
	//4字节之后存放的内容是消息体
	copy(buf[HeadLength:], data)

	return buf, nil
}

// 大端模式将int表位bytes
// PacketEncode packet data length to bytes(Big end)
func intToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}
