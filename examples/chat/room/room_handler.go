package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/examples/chat/protocol"
	clog "github.com/cherry-game/cherry/logger"
	chandler "github.com/cherry-game/cherry/net/handler"
	csession "github.com/cherry-game/cherry/net/session"
)

type roomHandler struct {
	chandler.Handler
}

func (h *roomHandler) Name() string {
	return "roomHandler"
}

func (h *roomHandler) OnInit() {
	h.AddLocal("syncMessage", h.syncMessage)
	csession.AddOnCreateListener(h.disconnected)
}

func (h *roomHandler) syncMessage(s *csession.Session, req *protocol.SyncMessage) {
	// Send an RPC to master server to stats
	stats(s, s.UID())

	rsp := &protocol.WriteResponses{}
	writeCode := cherry.RequestRemote("log-1", "log.logHandler.write", req, rsp)
	clog.Infof("--------> write code = %d, response = %+v", writeCode, rsp)

	// Sync message to all members in this room
	err := group.Broadcast("onMessage", req)
	if err != nil {
		clog.Error(err)
	}
}

func (h *roomHandler) disconnected(session *csession.Session) (next bool) {
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
