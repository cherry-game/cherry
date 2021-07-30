package cherryCommand

import (
	"encoding/json"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryMessage "github.com/cherry-game/cherry/net/message"
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
		"code":   200,
		"sys":    h.sysData,
		"routes": cherryMessage.GetDictionary(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		session.Warnf("data marshal error. error = %s", err)
		return
	}

	bytes, err := h.PacketEncode(cherryPacket.Handshake, jsonData)
	if err != nil {
		session.Warnf("handshake packet encode error. error = %s", err)
		return
	}

	session.SetState(cherrySession.WaitAck)
	err = session.SendRaw(bytes)
	if err != nil {
		cherryLogger.Error(err)
	}

	session.Debugf("request handshake. data[%v]", data)
}
