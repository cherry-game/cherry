package cherryCommand

import (
	facade "github.com/cherry-game/cherry/facade"
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
		session.Warnf("heartbeat packet encode error. error = %s", err)
		return
	}

	err = session.SendRaw(bytes)
	if err != nil {
		session.Warnf("send heartbeat raw data error. error = %s", err)
		return
	}

	session.Debug("request heartbeat.")
}
