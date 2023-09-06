package pomeloPacket

import (
	"bytes"

	cerr "github.com/cherry-game/cherry/error"
)

type (
	Type = byte // packet type
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

func TypeName(t Type) string {
	return packetTypes[t]
}

func InvalidType(t Type) bool {
	return t < Handshake || t > Kick
}

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, error) {
	if len(header) != HeadLength {
		return 0, cerr.PacketInvalidHeader
	}

	typ := header[0]
	if InvalidType(typ) {
		return 0, cerr.PacketWrongType
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, cerr.PacketSizeExceed
	}

	return size, nil
}

// BytesToInt Decode packet data length byte to int(Big end)
func BytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// IntToBytes Encode packet data length to bytes(Big end)
func IntToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}

func forward(buf *bytes.Buffer) (int, Type, error) {
	header := buf.Next(HeadLength)

	typ := header[0]
	if InvalidType(typ) {
		return 0, None, cerr.PacketSizeExceed
	}

	// get 2,3,4 byte
	size := BytesToInt(header[1:])

	// packet length limitation
	if size > MaxPacketSize {
		return 0, None, cerr.PacketSizeExceed
	}

	return size, typ, nil
}
