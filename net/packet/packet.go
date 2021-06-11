package cherryPacket

import (
	"fmt"
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
