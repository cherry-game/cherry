package chat

import (
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"sync/atomic"
)

type (
	UserHandler struct {
		cherryHandler.Handler
		nextId int64
	}
)

func (u *UserHandler) Name() string {
	return "userHandler"
}

func (u *UserHandler) OnInit() {
	u.RegisterLocal("login", u.login)
	u.RegisterLocal("testLogin", u.testLogin)
}

func (u *UserHandler) testLogin(_ *cherrySession.Session, _ *cherryMessage.Message, _ *LoginRequest) error {
	return nil
}

func (u *UserHandler) login(session *cherrySession.Session, msg *cherryMessage.Message, req *LoginRequest) error {
	atomic.AddInt64(&u.nextId, 1)

	uid := u.nextId
	request := &NewUserRequest{
		Nickname: req.Nickname,
		GateUid:  uid,
	}

	session.Debug(request)

	if err := NewUser(session, request); err != nil {
		return err
	}

	return session.Response(msg.ID, &LoginResponse{})
}
