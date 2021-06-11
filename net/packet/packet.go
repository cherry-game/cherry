package cherryPacket

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
)

type (
	// Decoder
	Decoder interface {
		Decode(data []byte) ([]*Packet, error)
	}

	// Encoder
	Encoder interface {
		Encode(typ byte, buf []byte) ([]byte, error)
	}
)

// Packet 网络消息包结构
type Packet struct {
	Type   byte   // 包类型
	Length int    // 包长度
	Data   []byte // 数据内容 message
}

// String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("Packet Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, byte, error) {
	if len(header) != HeadLength {
		return 0, 0x00, cherryError.PacketInvalidHeader
	}
	typ := header[0]
	if typ < Handshake || typ > Kick {
		return 0, None, cherryError.PacketWrongType
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, 0x00, cherryError.PacketSizeExceed
	}

	return size, typ, nil
}
