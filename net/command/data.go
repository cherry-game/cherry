package cherryCommand

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type (
	Data struct {
		post PostHandler
	}

	PostHandler func(session *cherrySession.Session, msg *cherryMessage.Message)
)

func NewData(postHandler PostHandler) *Data {
	data := &Data{
		post: postHandler,
	}
	return data
}

func (h *Data) GetType() cherryPacket.Type {
	return cherryPacket.Data
}

func (h *Data) Do(session *cherrySession.Session, packet facade.IPacket) {
	if session.State() != cherrySession.Working {
		session.Warnf("state is not working. state[%d]", session.State())
		return
	}

	msg, err := cherryMessage.Decode(packet.Data())
	if err != nil {
		session.Warnf("packet decode error. data[%s], error[%s].", packet.Data(), err)
		return
	}

	h.post(session, msg)
}
