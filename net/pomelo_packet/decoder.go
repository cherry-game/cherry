package cherryPomeloPacket

import (
	"bytes"
	"errors"
	"github.com/cherry-game/cherry/interfaces"
)

// ErrWrongPomeloPacketType represents a wrong packet type.
var ErrWrongPomeloPacketType = errors.New("wrong packet type")

type Decoder struct {
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (p *Decoder) Decode(data []byte) ([]*cherryInterfaces.Packet, error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(data)

	var (
		packets []*cherryInterfaces.Packet
		err     error
	)
	// check length
	if buf.Len() < HeadLength {
		return nil, nil
	}

	size, typ, err := p.forward(buf)
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		pkg := &cherryInterfaces.Packet{Type: typ, Length: size, Data: buf.Next(size)}
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

func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

func (p *Decoder) forward(buf *bytes.Buffer) (int, byte, error) {
	header := buf.Next(HeadLength) //读取4byte
	typ := header[0]               //取取消息类型
	if typ < Handshake || typ > Kick {
		return 0, 0x00, ErrWrongPomeloPacketType
	}
	size := bytesToInt(header[1:]) //获了消息长度

	// packet length limitation
	if size > MaxPacketSize {
		return 0, 0x00, ErrPacketSizeExcced
	}
	return size, typ, nil
}
