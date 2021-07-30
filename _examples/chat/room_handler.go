package main

import (
	"fmt"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type roomHandler struct {
	cherryHandler.Handler
	group *cherrySession.Group
}

func (h *roomHandler) Name() string {
	return "roomHandler"
}

func (h *roomHandler) OnInit() {
	h.group = cherrySession.NewGroup("all-users")

	h.AddLocal("joinRoom", h.joinRoom)
	h.AddLocal("syncMessage", h.syncMessage)

	cherrySession.AddOnCloseListener(h.disconnected)
}

func (h *roomHandler) joinRoom(session *cherrySession.Session, _ *cherryMessage.Message, req *joinRoomRequest) error {
	broadcast := &newUserBroadcast{
		Content: fmt.Sprintf("user join: %v", req.Nickname),
	}

	if err := h.group.Broadcast("onNewUser", broadcast); err != nil {
		return err
	}

	return h.group.Add(session)
}

func (h *roomHandler) syncMessage(s *cherrySession.Session, _ *cherryMessage.Message, req *syncMessage) error {
	// Send an RPC to master server to stats
	if err := stats(s, s.UID()); err != nil {
		return err
	}

	// Sync message to all members in this room
	return h.group.Broadcast("onMessage", req)
}

func (h *roomHandler) disconnected(session *cherrySession.Session) (next bool) {
	if session.SID() < 1 {
		return true
	}

	if err := h.group.Leave(session); err != nil {
		session.Debugf("Remove user from group failed. err=%s", err)
		return true
	}

	session.Debug("user session disconnected")
	return true
}
