package cherryCommand

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type Heartbeat struct {
	facade.IApplication
}

func NewHeartbeat(app facade.IApplication) *Heartbeat {
	return &Heartbeat{
		IApplication: app,
	}
}

func (h *Heartbeat) GetType() cherryPacket.Type {
	return cherryPacket.Heartbeat
}

func (h *Heartbeat) Do(session *cherrySession.Session, _ facade.IPacket) {
	bytes, err := h.PacketEncode(cherryPacket.Heartbeat, nil)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	err = session.SendRaw(bytes)
	if err != nil {
		cherryLogger.Error(err)
	}
	session.Debug("request heartbeat.")
}
