package cherryCommand

import (
	"encoding/json"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
	"go.uber.org/zap/zapcore"
	"time"
)

type Handshake struct {
	cfacade.IApplication
	sysData   map[string]interface{}
	heartbeat time.Duration
}

func (h *Handshake) PacketType() cpacket.Type {
	return cpacket.Handshake
}

func NewHandshake(app cfacade.IApplication, sysData map[string]interface{}) *Handshake {
	return &Handshake{
		IApplication: app,
		sysData:      sysData,
	}
}

func (h *Handshake) Do(session *csession.Session, _ cfacade.IPacket) {
	data := map[string]interface{}{
		"code": 200,
		"sys":  h.sysData,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		clog.Warn(err)
		return
	}

	bytes, err := h.PacketEncode(cpacket.Handshake, jsonData)
	if err != nil {
		clog.Warn(err)
		return
	}

	session.SetState(csession.WaitAck)
	session.SendRaw(bytes)

	if clog.LogLevel(zapcore.DebugLevel) {
		session.Debugf("request handshake. [sid = %s, address = %s, data = %v]",
			session.SID(),
			session.RemoteAddress(),
			data,
		)
	}
}
