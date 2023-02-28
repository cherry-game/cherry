package main

import (
	"fmt"
	cherryCode "github.com/cherry-game/cherry/code"
	"github.com/cherry-game/cherry/examples/demo_chat/protocol"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
)

type (
	actorWrite struct {
		cactor.Base
	}
)

func (p *actorWrite) AliasID() string {
	return "write"
}

func (p *actorWrite) OnInit() {
	p.Remote().Register("write1", p.write1)
}

func (p *actorWrite) write1(req *protocol.SyncMessage) (*protocol.WriteResponse, int32) {
	clog.Debugf("write log. message = %+v", req)
	rsp := &protocol.WriteResponse{
		Result: true,
	}

	fmt.Println("write1--->", req.PacketSpendTime())

	return rsp, cherryCode.OK
}
