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
		Encode(typ Type, buf []byte) ([]byte, error)
	}
)

// Packet 网络消息包结构
type Packet struct {
	Type   Type   // 包类型
	Length int    // 包长度
	Data   []byte // 数据内容 message
}

// String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("packet type: %s, length: %d, data: %s", p.Type.String(), p.Length, string(p.Data))
}

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, Type, error) {
	if len(header) != HeadLength {
		return 0, None, cherryError.PacketInvalidHeader
	}
	typ := header[0]
	if invalidType(Type(typ)) {
		return 0, None, cherryError.PacketWrongType
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, None, cherryError.PacketSizeExceed
	}

	return size, Type(typ), nil
}
