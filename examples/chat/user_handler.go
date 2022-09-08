package main

import (
	"context"
	chandler "github.com/cherry-game/cherry/net/handler"
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
}

func (h *userHandler) login(ctx context.Context, session *csession.Session, req *loginRequest) {
	session.Debugf("nickname = %s", req.Nickname)
	if err := newUser(session, req.Nickname); err != nil {
		return
	}

	session.Response(ctx, &loginResponse{})
}
