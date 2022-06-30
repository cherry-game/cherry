package main

import (
	"context"
	chandler "github.com/cherry-game/cherry/net/handler"
	cmsg "github.com/cherry-game/cherry/net/message"
	csession "github.com/cherry-game/cherry/net/session"
)

type (
	userHandler struct {
		chandler.Handler
	}
)

func (h *userHandler) Name() string {
	return "userHandler"
}

func (h *userHandler) OnInit() {
	h.AddLocal("login", h.login)

	csession.AddOnCloseListener(disconnect)

	h.AddBeforeFilter(func(ctx context.Context, session *csession.Session, message *cmsg.Message) bool {
		//if session.IsBind() == false && message.Route != "game.userHandler.login" {
		//	//登录后，才能发送后续消息
		//	session.Kick(fmt.Sprintf("kick %s : not login", session.String()), true)
		//	return false
		//}
		//cherryContext.GetMessageId(ctx)
		return true
	})
}

func (h *userHandler) login(ctx context.Context, session *csession.Session, req *loginRequest) {
	session.Debugf("nickname = %s", req.Nickname)
	if err := newUser(session, req.Nickname); err != nil {
		return
	}

	session.Response(ctx, &loginResponse{})
}
