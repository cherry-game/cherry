package main

import (
	cherryCode "github.com/cherry-game/cherry/code"
	"github.com/cherry-game/cherry/examples/demo_chat/protocol"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	"time"
)

type (
	ActorLog struct {
		pomelo.ActorBase
	}
)

func (*ActorLog) AliasID() string {
	return "log"
}

func (p *ActorLog) OnInit() {
	p.Remote().Register("write", p.write)

	p.Timer().Add(time.Second, p.secondTimer)
	p.Timer().AddFixedHour(14, 54, 0, p.fixedHourTimer)
	p.Timer().AddFixedMinute(54, 19, p.fixedMinuteTimer)
}

func (p *ActorLog) write(req *protocol.SyncMessage) (*protocol.WriteResponse, int32) {
	clog.Debugf("write log. message = %+v", req)
	rsp := &protocol.WriteResponse{
		Result: true,
	}

	clog.Debugf("write---> %d(nano second)", req.PacketSpendTime())

	return rsp, cherryCode.OK
}

func (p *ActorLog) secondTimer() {
	clog.Info("secondTimer")
}

func (p *ActorLog) fixedHourTimer() {
	clog.Info("fixedHourTimer")
}

func (p *ActorLog) fixedMinuteTimer() {
	clog.Info("fixedMinuteTimer")
}
