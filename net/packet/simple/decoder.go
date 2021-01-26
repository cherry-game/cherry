package cherryPacketSimple

import (
	"github.com/cherry-game/cherry/interfaces"
)

type Decoder struct {
}

func (s *Decoder) Decode(data []byte) ([]*cherryInterfaces.Packet, error) {
	return nil, nil
}
