package cherryPacket

import cherryError "github.com/cherry-game/cherry/error"

const (
	// None
	None = 0x00

	// Handshake represents a handshake: request(client) <====> handshake response(server)
	Handshake = 0x01

	// HandshakeAck represents a handshake ack from client to server
	HandshakeAck = 0x02

	// Heartbeat represents a heartbeat
	Heartbeat = 0x03

	// Settings represents a common data packet
	Data = 0x04

	// Kick represents a kick off packet
	Kick = 0x05 // disconnect message from server
)

var HeadLength = 4          // 4 bytes
var MaxPacketSize = 1 << 24 // 16mb

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, byte, error) {
	if len(header) != HeadLength {
		return 0, 0x00, cherryError.PacketInvalidHeader
	}
	typ := header[0]
	if typ < Handshake || typ > Kick {
		return 0, 0x00, cherryError.PacketWrongType
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, 0x00, cherryError.PacketSizeExceed
	}

	return size, typ, nil
}
