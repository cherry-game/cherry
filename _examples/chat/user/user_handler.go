package user

import (
	"github.com/cherry-game/cherry/_examples/chat/proto"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
)

type (
	Handler struct {
		cherryHandler.Handler
	}
)

func (h *Handler) Name() string {
	return "userHandler"
}

func (h *Handler) OnInit() {
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

func (h *Handler) testLogin(_ *cherrySession.Session, _ *cherryMessage.Message, _ *proto.LoginRequest) error {
	return nil
}

func (h *Handler) login(session *cherrySession.Session, msg *cherryMessage.Message, req *proto.LoginRequest) error {
	session.Debugf("nickname = %s", req.Nickname)

	if err := NewUser(session, req.Nickname); err != nil {
		return err
	}

	return session.Response(msg.ID, &proto.LoginResponse{})
}
