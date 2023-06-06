package simple

import (
	"bytes"
	"encoding/binary"
	cerr "github.com/cherry-game/cherry/error"
	"io"
	"net"
)

var (
	NoneMessage        = Message{} // none message
	headLength         = 8         // ID uint32(4 bytes) +  DataLen uint32(4 bytes)
	dataLength  uint32 = 4096      // data length
)

type Message struct {
	MID  uint32
	Len  uint32
	Data []byte
}

func ReadMessage(conn net.Conn) (Message, bool, error) {
	header, err := io.ReadAll(io.LimitReader(conn, int64(headLength)))
	if err != nil {
		return NoneMessage, true, err
	}

	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return NoneMessage, true, cerr.PacketConnectClosed
	}

	msg, err := parseHeader(header)
	if err != nil {
		return NoneMessage, true, err
	}

	msgData, err := io.ReadAll(io.LimitReader(conn, int64(msg.Len)))
	if err != nil {
		return NoneMessage, true, err
	}

	msg.Data = msgData

	return msg, false, nil
}

func parseHeader(header []byte) (Message, error) {
	msg := Message{}

	if len(header) != headLength {
		return msg, cerr.PacketInvalidHeader
	}

	bytesReader := bytes.NewReader(header)
	err := binary.Read(bytesReader, endian, &msg.MID)
	if err != nil {
		return msg, err
	}

	err = binary.Read(bytesReader, endian, &msg.Len)
	if err != nil {
		return msg, err
	}

	if msg.Len > dataLength {
		return msg, cerr.PacketSizeExceed
	}

	return msg, nil
}

func pack(mid uint32, data []byte) ([]byte, error) {
	pkg := bytes.NewBuffer([]byte{})
	binary.Write(pkg, endian, mid)
	binary.Write(pkg, endian, uint32(len(data)))
	binary.Write(pkg, endian, data)
	return pkg.Bytes(), nil
}
