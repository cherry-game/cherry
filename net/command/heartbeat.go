package cherryCommand

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
	"go.uber.org/zap/zapcore"
)

type Heartbeat struct {
	cfacade.IApplication
}

func NewHeartbeat(app cfacade.IApplication) *Heartbeat {
	return &Heartbeat{
		IApplication: app,
	}
}

func (h *Heartbeat) PacketType() cpacket.Type {
	return cpacket.Heartbeat
}

func (h *Heartbeat) Do(session *csession.Session, _ cfacade.IPacket) {
	bytes, err := h.PacketEncode(cpacket.Heartbeat, nil)
	if err != nil {
		clog.Warn(err)
		return
	}

	session.SendRaw(bytes)
	if clog.LogLevel(zapcore.DebugLevel) {
		session.Debug("response heartbeat.")
	}
}
