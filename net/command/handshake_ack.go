package cherryCommand

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type HandshakeACK struct {
}

func NewHandshakeACK() *HandshakeACK {
	return &HandshakeACK{}
}

func (h *HandshakeACK) GetType() cherryPacket.Type {
	return cherryPacket.HandshakeAck
}

func (h *HandshakeACK) Do(session *cherrySession.Session, _ facade.IPacket) {
	session.SetState(cherrySession.Working)
	session.Debugf("request handshakeACK. [sid = %s, address = %s]",
		session.SID(),
		session.RemoteAddress(),
	)
}
