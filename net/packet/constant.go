package cherryPacket

type Type byte

var types = map[Type]string{
	None:         "None",
	Handshake:    "Handshake",
	HandshakeAck: "HandshakeAck",
	Heartbeat:    "Heartbeat",
	Data:         "Data",
	Kick:         "Kick",
}

func (t *Type) String() string {
	return types[*t]
}

func invalidType(t Type) bool {
	return t < Handshake || t > Kick
}

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

var HeadLength = 4          // 4 bytes
var MaxPacketSize = 1 << 24 // 16mb
