package main

import (
	"context"
	"github.com/cherry-game/cherry/_examples/test_sample1/internal/protocols"
	ctime "github.com/cherry-game/cherry/extend/time"
	clog "github.com/cherry-game/cherry/logger"
	chandler "github.com/cherry-game/cherry/net/handler"
	csession "github.com/cherry-game/cherry/net/session"
)

type ActorHandler struct {
	chandler.Handler
}

func (h *ActorHandler) Name() string {
	return "actorHandler"
}

func (h *ActorHandler) OnInit() {
	h.AddLocal("select", h.selectActor)
	h.AddLocal("create", h.creatActor)
}

// selectActor 查询角色
func (h *ActorHandler) selectActor(ctx context.Context, session *csession.Session, req *protocols.ActorSelectRequest) error {
	session.Info("call selectActor()")

	now := ctime.Now().ToMillisecond()
	clog.Debugf("now = %d,  clientTime = %d, result = %d", now, req.Time, now-int64(req.Time))

	rsp := &protocols.ActorSelectResponse{ActorName: "hello"}
	session.Response(ctx, rsp)

	session.Push("onBalance", &protocols.LoginResponse{Uid: 444})
	//session.Kick(&protocols.LoginResponse{Uid: 444}, true)

	return nil
}

func (h *ActorHandler) creatActor(ctx context.Context, session *csession.Session, req *protocols.ActorCreateRequest) error {
	session.Info("call creatActor()")
	return nil
}
