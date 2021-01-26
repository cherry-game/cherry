package cherryPacketPomelo

import (
	"errors"
)

const (
	Inited = iota
	WaitAck
	Working
	Closed
)

var SessionStatus = map[int]string{
	Inited:  "inited",
	WaitAck: "wait_ack",
	Working: "working",
	Closed:  "closed",
}

const (
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

const (
	HeadLength    = 4       //4 bytes
	MaxPacketSize = 1 << 24 //16mb
)

//var error
var (
	ErrWrongPacketType = errors.New("wrong packet type")

	// ErrPacketSizeExcced is the error used for encode/decode.
	ErrPacketSizeExcced = errors.New("codec: packet size exceed")
)
