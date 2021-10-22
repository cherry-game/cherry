package main

import (
	"github.com/cherry-game/cherry/_examples/test_sample1/internal/protocols"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	ch "github.com/cherry-game/cherry/net/handler"
	cm "github.com/cherry-game/cherry/net/message"
	cs "github.com/cherry-game/cherry/net/session"
)

type ActorHandler struct {
	ch.Handler
}

func (h *ActorHandler) Name() string {
	return "actorHandler"
}

func (h *ActorHandler) OnInit() {
	h.AddLocal("select", h.selectActor)
	h.AddLocal("create", h.creatActor)
}

// selectActor 查询角色
func (h *ActorHandler) selectActor(session *cs.Session, msg *cm.Message, req *protocols.ActorSelectRequest) error {
	session.Info("call selectActor()")

	now := cherryTime.Now().ToMillisecond()
	cherryLogger.Debugf("now = %d,  clientTime = %d, result = %d", now, req.Time, now-int64(req.Time))

	rsp := &protocols.ActorSelectResponse{ActorName: "hello"}
	session.ResponseMID(msg.ID, rsp)

	session.Push("onBalance", &protocols.LoginResponse{Uid: 444})
	//session.Kick(&protocols.LoginResponse{Uid: 444}, true)

	return nil
}

func (h *ActorHandler) creatActor(session *cs.Session, msg *cm.Message, req *protocols.ActorCreateRequest) error {
	return nil
}
