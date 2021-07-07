package cherryPacket

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
)

const (
	// None
	None Type = 0x00

	// Handshake represents a handshake: request(client) <====> handshake response(server)
	Handshake Type = 0x01

	// HandshakeAck represents a handshake ack from client to server
	HandshakeAck Type = 0x02

	// Heartbeat represents a heartbeat
	Heartbeat Type = 0x03

	// settings represents a common data packet
	Data Type = 0x04

	// Kick represents a kick off packet
	Kick Type = 0x05 // disconnect message from server
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
		return 0, None, cherryError.PacketWrongType
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, None, cherryError.PacketSizeExceed
	}

	return size, typ, nil
}
