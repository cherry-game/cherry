package main

import (
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
	h.AddLocal("testLogin", h.testLogin)

	h.AddOnClose(disconnect)

	h.AddBeforeFilter(func(session *cherrySession.Session, message *cherryMessage.Message) bool {
		if session.IsBind() == false && message.Route != "gate.userHandler.login" {
			//登录后，才能发送后续消息
			session.Kick("not login", true)
			return false
		}
		return true
	})
}

func (h *userHandler) testLogin(_ *cherrySession.Session, _ *cherryMessage.Message, _ *loginRequest) error {
	return nil
}

func (h *userHandler) login(session *cherrySession.Session, msg *cherryMessage.Message, req *loginRequest) error {
	session.Debugf("nickname = %s", req.Nickname)

	if err := newUser(session, req.Nickname); err != nil {
		return err
	}

	return session.Response(msg.ID, &loginResponse{})
}
