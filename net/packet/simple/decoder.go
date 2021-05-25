package cherryPacketSimple

import (
	"github.com/cherry-game/cherry/net/packet"
)

type Decoder struct {
}

func (s *Decoder) Decode(data []byte) ([]*cherryPacket.Packet, error) {
	return nil, nil
}
