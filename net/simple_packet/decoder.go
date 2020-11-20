package cherrySimplePacket

import (
	"github.com/phantacix/cherry/interfaces"
)

type Decoder struct {
}

func (s *Decoder) Decode(data []byte) ([]*cherryInterfaces.Packet, error) {
	return nil, nil
}
