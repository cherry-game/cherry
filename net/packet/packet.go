package cherryPacket

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
)

const (
	None         Type = 0x00 // None
	Handshake    Type = 0x01 // Handshake represents a handshake: request(client) <====> handshake response(server)
	HandshakeAck Type = 0x02 // HandshakeAck represents a handshake ack from client to server
	Heartbeat    Type = 0x03 // Heartbeat represents a heartbeat
	Data         Type = 0x04 // settings represents a common data packet
	Kick         Type = 0x05 // Kick represents a kick off packet
)

var (
	HeadLength    = 4       // 4 bytes
	MaxPacketSize = 1 << 24 // 16mb

	packetTypes = map[Type]string{
		None:         "None",
		Handshake:    "Handshake",
		HandshakeAck: "HandshakeAck",
		Heartbeat:    "Heartbeat",
		Data:         "Data",
		Kick:         "Kick",
	}
)

type (
	Type = byte

	Packet struct {
		typ  Type
		len  int
		data []byte
	}
)

func TypeName(t Type) string {
	return packetTypes[t]
}

func InvalidType(t Type) bool {
	return t < Handshake || t > Kick
}

func (p *Packet) Type() Type {
	return p.typ
}

func (p *Packet) Len() int {
	return p.len
}

func (p *Packet) Data() []byte {
	return p.data
}

func (p *Packet) SetData(data []byte) {
	p.data = data
}

// String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("packet type: %s, length: %d, data: %s", TypeName(p.typ), p.len, string(p.data))
}

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, Type, error) {
	if len(header) != HeadLength {
		return 0, None, cherryError.PacketInvalidHeader
	}

	typ := header[0]
	if InvalidType(typ) {
		return 0, None, cherryError.Errorf("wrong packet type. typ = %d, header = %v", typ, header)
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, None, cherryError.PacketSizeExceed
	}

	return size, typ, nil
}
