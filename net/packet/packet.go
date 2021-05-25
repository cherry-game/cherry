package cherryPacket

import "fmt"

type (
	// Decoder 网络消息包解码接口
	Decoder interface {
		Decode(data []byte) ([]*Packet, error)
	}

	// Encoder 网络消息包编码接口
	Encoder interface {
		Encode(typ byte, buf []byte) ([]byte, error)
	}

	// Packet 单个网络消息包结构
	Packet struct {
		Type   byte   // 包类型
		Length int    // 包长度
		Data   []byte // 数据内容
	}
)

// New create a Packet instance.
func New() *Packet {
	return &Packet{}
}

// String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Settings: %s", p.Type, p.Length, string(p.Data))
}
