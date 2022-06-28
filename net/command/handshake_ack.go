package cherryCommand

import (
	cfacade "github.com/cherry-game/cherry/facade"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
)

type HandshakeACK struct {
}

func NewHandshakeACK() *HandshakeACK {
	return &HandshakeACK{}
}

func (h *HandshakeACK) PacketType() cpacket.Type {
	return cpacket.HandshakeAck
}

func (h *HandshakeACK) Do(session *csession.Session, _ cfacade.IPacket) {
	session.SetState(csession.Working)
	session.Debugf("request handshakeACK. [sid = %s, address = %s]",
		session.SID(),
		session.RemoteAddress(),
	)
}
