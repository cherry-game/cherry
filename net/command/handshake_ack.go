package cherryCommand

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
	"go.uber.org/zap/zapcore"
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

	if clog.LogLevel(zapcore.DebugLevel) {
		session.Debugf("request handshakeACK. [sid = %s, address = %s]",
			session.SID(),
			session.RemoteAddress(),
		)
	}
}
