package cherryCommand

import (
	"context"
	cfacade "github.com/cherry-game/cherry/facade"
	ccontext "github.com/cherry-game/cherry/net/context"
	cmsg "github.com/cherry-game/cherry/net/message"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
	"time"
)

type (
	Data struct {
		cfacade.IApplication
		localMessage   ProcessMessage
		forwardMessage ProcessMessage
	}

	ProcessMessage func(ctx context.Context, session *csession.Session, msg *cmsg.Message)
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

	ctx := ccontext.New()
	ctx = ccontext.Add(ctx, ccontext.BuildPacketTimeKey, time.Now().UnixMilli())
	ctx = ccontext.Add(ctx, ccontext.MessageIdKey, msg.ID)

	if msg.RouteInfo().NodeType() == h.NodeType() {
		h.localMessage(ctx, session, msg)
	} else {
		h.forwardMessage(ctx, session, msg)
	}
}
