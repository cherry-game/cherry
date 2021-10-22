package main

import (
	"context"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
)

type (
	userHandler struct {
		cherryHandler.Handler
	}
)

func (h *userHandler) Name() string {
	return "userHandler"
}

func (h *userHandler) OnInit() {
	h.AddLocal("login", h.login)

	cherrySession.AddOnCloseListener(disconnect)

	h.AddBeforeFilter(func(ctx context.Context, session *cherrySession.Session, message *cherryMessage.Message) bool {
		//if session.IsBind() == false && message.Route != "game.userHandler.login" {
		//	//登录后，才能发送后续消息
		//	session.Kick(fmt.Sprintf("kick %s : not login", session.String()), true)
		//	return false
		//}
		//cherryContext.GetMessageId(ctx)
		return true
	})
}

func (h *userHandler) login(ctx context.Context, session *cherrySession.Session, req *loginRequest) {
	session.Debugf("nickname = %s", req.Nickname)
	if err := newUser(session, req.Nickname); err != nil {
		return
	}

	session.Response(ctx, &loginResponse{})
}
