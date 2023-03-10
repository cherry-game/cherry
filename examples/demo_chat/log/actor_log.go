package main

import (
	cherryCode "github.com/cherry-game/cherry/code"
	"github.com/cherry-game/cherry/examples/demo_chat/protocol"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
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
}

func (p *ActorLog) write(req *protocol.SyncMessage) (*protocol.WriteResponse, int32) {
	clog.Debugf("write log. message = %+v", req)
	rsp := &protocol.WriteResponse{
		Result: true,
	}

	clog.Debugf("write---> %d(nano second)", req.PacketSpendTime())

	return rsp, cherryCode.OK
}
