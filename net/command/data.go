package cherryCommand

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type (
	Data struct {
		facade.IApplication
		localHandler   ProcessHandler
		forwardHandler ProcessHandler
	}

	ProcessHandler func(session *cherrySession.Session, msg *cherryMessage.Message)
)

func NewData(app facade.IApplication, localHandler ProcessHandler, forwardHandler ProcessHandler) *Data {
	data := &Data{
		IApplication:   app,
		localHandler:   localHandler,
		forwardHandler: forwardHandler,
	}
	return data
}

func (h *Data) GetType() cherryPacket.Type {
	return cherryPacket.Data
}

func (h *Data) Do(session *cherrySession.Session, packet facade.IPacket) {
	if session.State() != cherrySession.Working {
		session.Warnf("state is not working. [state = %d]", session.State())
		return
	}

	msg, err := cherryMessage.Decode(packet.Data())
	if err != nil {
		session.Warnf("packet decode error. [data = %s, error = %s]", packet.Data(), err)
		return
	}

	err = msg.ParseRoute()
	if err != nil {
		session.Warnf("packet decode route error. [data = %s, error = %s]", packet.Data(), err)
		return
	}

	if msg.RouteInfo().NodeType() == h.NodeType() {
		h.localHandler(session, msg)
	} else {
		h.forwardHandler(session, msg)
	}
}
