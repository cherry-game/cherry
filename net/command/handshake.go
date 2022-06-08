package cherryCommand

import (
	"encoding/json"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"time"
)

type Handshake struct {
	facade.IApplication
	sysData   map[string]interface{}
	heartbeat time.Duration
}

func (h *Handshake) GetType() cherryPacket.Type {
	return cherryPacket.Handshake
}

func NewHandshake(app facade.IApplication, sysData map[string]interface{}) *Handshake {
	return &Handshake{
		IApplication: app,
		sysData:      sysData,
	}
}

func (h *Handshake) Do(session *cherrySession.Session, _ facade.IPacket) {
	data := map[string]interface{}{
		"code": 200,
		"sys":  h.sysData,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	bytes, err := h.PacketEncode(cherryPacket.Handshake, jsonData)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	session.SetState(cherrySession.WaitAck)
	session.SendRaw(bytes)

	session.Debugf("request handshake. [sid = %s, address = %s, data = %v]",
		session.SID(),
		session.RemoteAddress(),
		data,
	)
}
