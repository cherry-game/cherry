package room

import (
	"fmt"
	"github.com/cherry-game/cherry/_examples/chat/proto"
	"github.com/cherry-game/cherry/_examples/chat/user"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
)

type Handler struct {
	cherryHandler.Handler
	group *cherrySession.Group
}

func (h *Handler) Name() string {
	return "roomHandler"
}

func (h *Handler) OnInit() {
	h.group = cherrySession.NewGroup("all-users")

	h.RegisterLocal("joinRoom", h.joinRoom)
	h.RegisterLocal("syncMessage", h.syncMessage)

	h.AddOnClose(h.disconnected)
}

func (h *Handler) joinRoom(session *cherrySession.Session, _ *cherryMessage.Message, req *proto.JoinRoomRequest) error {
	broadcast := &proto.NewUserBroadcast{
		Content: fmt.Sprintf("user join: %v", req.Nickname),
	}

	if err := h.group.Broadcast("onNewUser", broadcast); err != nil {
		return err
	}

	return h.group.Add(session)
}

func (h *Handler) syncMessage(s *cherrySession.Session, _ *cherryMessage.Message, req *proto.SyncMessage) error {
	// Send an RPC to master server to stats
	if err := user.Stats(s, s.UID()); err != nil {
		return err
	}

	// Sync message to all members in this room
	return h.group.Broadcast("onMessage", req)
}

func (h *Handler) disconnected(session *cherrySession.Session) (next bool) {
	if session.UID() < 1 {
		return true
	}

	if err := h.group.Leave(session); err != nil {
		session.Debugf("Remove user from group failed. err=%s", err)
		return true
	}

	session.Debug("User session disconnected")
	return true
}
