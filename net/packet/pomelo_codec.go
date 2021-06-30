package cherryPacket

import (
	"bytes"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
)

type PomeloCodec struct {
}

func NewPomeloCodec() *PomeloCodec {
	return &PomeloCodec{}
}

func (p *PomeloCodec) PacketDecode(data []byte) ([]cherryFacade.IPacket, error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(data)

	var (
		packets []cherryFacade.IPacket
		err     error
	)

	// check length
	if buf.Len() < HeadLength {
		return nil, err
	}

	size, typ, err := p.forward(buf)
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		pkg := &Packet{
			typ:  typ,
			len:  size,
			data: buf.Next(size),
		}

		packets = append(packets, pkg)

		if buf.Len() < HeadLength {
			break
		}

		size, typ, err = p.forward(buf)
		if err != nil {
			return nil, err
		}
	}

	return packets, nil
}

func (p *PomeloCodec) forward(buf *bytes.Buffer) (int, Type, error) {
	header := buf.Next(HeadLength)

	typ := Type(header[0])

	if typ.InvalidType() {
		return 0, None, cherryError.PacketSizeExceed
	}

	// get 2,3,4 byte
	size := BytesToInt(header[1:])

	// packet length limitation
	if size > MaxPacketSize {
		return 0, None, cherryError.PacketSizeExceed
	}

	return size, typ, nil
}

// PacketEncode create a packet.Packet from  the raw bytes slice and then encode to network bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func (p *PomeloCodec) PacketEncode(typ byte, data []byte) ([]byte, error) {
	t := Type(typ)
	if t < Handshake || t > Kick {
		return nil, cherryError.PacketWrongType
	}

	if len(data) > MaxPacketSize {
		return nil, cherryError.PacketSizeExceed
	}

	pkg := &Packet{
		typ: t,
		len: len(data),
	}

	// header+body = 4 + len(body)
	buf := make([]byte, pkg.len+HeadLength)

	//第一个字节存放消息类型
	buf[0] = pkg.Type()

	//2~4 字节 存放消息长度
	copy(buf[1:HeadLength], IntToBytes(pkg.len))

	//4字节之后存放的内容是消息体
	copy(buf[HeadLength:], data)

	return buf, nil
}

// Decode packet data length byte to int(Big end)
func BytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// Encode packet data length to bytes(Big end)
func IntToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}
