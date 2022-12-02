package main

import (
	cherryCode "github.com/cherry-game/cherry/code"
	"github.com/cherry-game/cherry/examples/chat/protocol"
	clog "github.com/cherry-game/cherry/logger"
	ch "github.com/cherry-game/cherry/net/handler"
)

type (
	Handler struct {
		ch.Handler
	}
)

func (p *Handler) Name() string {
	return "logHandler"
}

func (p *Handler) OnInit() {
	p.AddRemote("write", p.write)
}

// ping 请求center是否响应
func (p *Handler) write(req *protocol.SyncMessage) (*protocol.WriteResponses, int32) {
	clog.Debugf("message = %+v", req)
	return &protocol.WriteResponses{Result: true}, cherryCode.OK
}
