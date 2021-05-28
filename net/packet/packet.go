package cherryPacket

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
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
	Data   []byte // 数据内容
}

// String represents the Packet's in text mode.
func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}

func GetNextMessage(conn net.Conn) ([]byte, error) {
	header, err := ioutil.ReadAll(io.LimitReader(conn, HeadLength))
	if err != nil {
		return nil, err
	}

	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, ErrPacketSizeExcced
	}

	msgSize, _, err := ParseHeader(header)
	if err != nil {
		return nil, err
	}

	msgData, err := ioutil.ReadAll(io.LimitReader(conn, int64(msgSize)))
	if err != nil {
		return nil, err
	}

	if len(msgData) < msgSize {
		return nil, ErrPacketSizeExcced
	}

	return append(header, msgData...), nil
}

func ParseHeader(header []byte) (int, byte, error) {
	if len(header) != HeadLength {
		return 0, None, ErrPacketSizeExcced
	}

	typ := header[0]
	if typ < Handshake || typ > Kick {
		return 0, None, ErrPacketSizeExcced
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, None, ErrPacketSizeExcced
	}

	return size, typ, nil
}
