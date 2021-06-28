package chat

import (
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"sync/atomic"
)

type (
	UserHandler struct {
		cherryHandler.Handler
		nextId int64
	}
)

func (u *UserHandler) Name() string {
	return "BindService"
}

func (u *UserHandler) OnInit() {
	u.RegisterLocal("Login", u.Login)
}

func (u *UserHandler) Login(session *cherrySession.Session, msg *LoginRequest) {
	atomic.AddInt64(&u.nextId, 1)

	//uid := u.nextId
	//request := &NewUserRequest{
	//	Nickname: msg.Nickname,
	//	GateUid:  uid,
	//}

	//session.Debug(request)

	//cherryLogger.Info(request)
	//if err := s.RPC("TopicService.NewUser", request); err != nil {
	//	return errors.Trace(err)
	//}
	//
	//session.Response(0,&LoginResponse{})
}
