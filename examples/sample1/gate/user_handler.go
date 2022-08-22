package main

import (
	"context"
	"fmt"
	"github.com/cherry-game/cherry/examples/sample1/constant"
	"github.com/cherry-game/cherry/examples/sample1/internal/protocols"
	cherrySnowflake "github.com/cherry-game/cherry/extend/snowflake"
	cherryFacade "github.com/cherry-game/cherry/facade"
	ch "github.com/cherry-game/cherry/net/handler"
	cm "github.com/cherry-game/cherry/net/message"
	cs "github.com/cherry-game/cherry/net/session"
)

type (
	UserHandler struct {
		ch.Handler
		discovery cherryFacade.IDiscovery
	}
)

const (
	firstHandlerName = "gate.userHandler.login"
)

func (h *UserHandler) Name() string {
	return "userHandler"
}

func (h *UserHandler) OnInit() {
	h.AddLocal("login", h.login)

	h.AddBeforeFilter(func(ctx context.Context, session *cs.Session, message *cm.Message) bool {
		if !session.IsBind() && message.Route != firstHandlerName {
			// session绑定后才能处理后续的消息
			session.Kick(fmt.Sprintf("kick %d : not login", session.SID()), true)
			return false
		}
		return true
	})

	cherrySnowflake.SetDefaultNode(0)
}

func (h *UserHandler) login(ctx context.Context, session *cs.Session, req *protocols.LoginRequest) error {
	// set server id
	session.Set(constant.GameServerId, req.ServerId)

	uid := cherrySnowflake.NextId()

	// uid bind session
	cs.Bind(session.SID(), uid)

	rsp := &protocols.LoginResponse{Uid: uid}
	session.Response(ctx, rsp)

	return nil
}
