package cherryCommand

import (
	cfacade "github.com/cherry-game/cherry/facade"
	cmsg "github.com/cherry-game/cherry/net/message"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
)

type (
	Data struct {
		cfacade.IApplication
		localMessage   ProcessMessage
		forwardMessage ProcessMessage
	}

	ProcessMessage func(session *csession.Session, msg *cmsg.Message)
)

func NewData(app cfacade.IApplication, localMessage ProcessMessage, forwardMessage ProcessMessage) *Data {
	return &Data{
		IApplication:   app,
		localMessage:   localMessage,
		forwardMessage: forwardMessage,
	}
}

func (h *Data) PacketType() cpacket.Type {
	return cpacket.Data
}

func (h *Data) Do(session *csession.Session, packet cfacade.IPacket) {
	if session.State() != csession.Working {
		session.Warnf("state is not working. [state = %d]", session.State())
		return
	}

	msg, err := cmsg.Decode(packet.Data())
	if err != nil {
		session.Warnf("packet decode error. [data = %s, error = %s]", packet.Data(), err)
		return
	}

	if err = msg.ParseRoute(); err != nil {
		session.Warnf("packet decode route error. [data = %s, error = %s]", packet.Data(), err)
		return
	}

	if msg.RouteInfo().NodeType() == h.NodeType() {
		h.localMessage(session, msg)
	} else {
		h.forwardMessage(session, msg)
	}
}
