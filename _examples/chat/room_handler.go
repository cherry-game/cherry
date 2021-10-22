package main

import (
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type roomHandler struct {
	cherryHandler.Handler
}

func (h *roomHandler) Name() string {
	return "roomHandler"
}

func (h *roomHandler) OnInit() {
	h.AddLocal("syncMessage", h.syncMessage)
	cherrySession.AddOnCreateListener(h.disconnected)
}

func (h *roomHandler) syncMessage(s *cherrySession.Session, req *syncMessage) error {
	// Send an RPC to master server to stats
	stats(s, s.UID())

	// Sync message to all members in this room
	return group.Broadcast("onMessage", req)
}

func (h *roomHandler) disconnected(session *cherrySession.Session) (next bool) {
	if session.UID() < 1 {
		return true
	}

	if err := group.Leave(session); err != nil {
		session.Debugf("Remove user from group failed. err=%s", err)
		return true
	}

	session.Debug("user session disconnected")
	return true
}
