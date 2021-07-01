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
	h.RegisterLocal("login", h.login)
	h.RegisterLocal("testLogin", h.testLogin)

	h.AddOnClose(Disconnect)

	h.AddBeforeFilter(func(executor *cherryHandler.MessageExecutor) bool {
		if executor.Session.IsBind() == false && executor.Msg.Route != "gate.userHandler.login" {
			//限制登陆后，才能发送后续的消息
			executor.Session.Kick("not login", true)
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
