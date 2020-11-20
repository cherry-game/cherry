package cherryInterfaces

import "fmt"

type PacketDecoder interface {
	Decode(data []byte) ([]*Packet, error)
}

type PacketEncoder interface {
	Encode(typ byte, buf []byte) ([]byte, error)
}

// Packet represents a network packet.
type Packet struct {
	Type   byte
	Length int
	Data   []byte
}

//New create a Packet instance.
func New() *Packet {
	return &Packet{}
}

//String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Settings: %s", p.Type, p.Length, string(p.Data))
}
