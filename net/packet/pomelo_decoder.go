package cherryPacket

import (
	"bytes"
	"github.com/cherry-game/cherry/error"
)

type PomeloDecoder struct {
}

func NewPomeloDecoder() *PomeloDecoder {
	return &PomeloDecoder{}
}

func (p *PomeloDecoder) Decode(data []byte) ([]*Packet, error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(data)

	var (
		packets []*Packet
		err     error
	)

	// check length
	if buf.Len() < HeadLength {
		return nil, err
	}

	size, typ, err := p.forward(buf)
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		pkg := &Packet{
			Type:   typ,
			Length: size,
			Data:   buf.Next(size),
		}

		packets = append(packets, pkg)

		if buf.Len() < HeadLength {
			break
		}

		size, typ, err = p.forward(buf)
		if err != nil {
			return nil, err
		}
	}

	return packets, nil
}

func (p *PomeloDecoder) forward(buf *bytes.Buffer) (int, Type, error) {
	header := buf.Next(HeadLength)
	typ := header[0]

	if invalidType(Type(typ)) {
		return 0, None, cherryError.PacketSizeExceed
	}

	// get 2,3,4 byte
	size := BytesToInt(header[1:])

	// packet length limitation
	if size > MaxPacketSize {
		return 0, None, cherryError.PacketSizeExceed
	}

	return size, Type(typ), nil
}

// Decode packet data length byte to int(Big end)
func BytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}
