package ops

import (
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	ch "github.com/cherry-game/cherry/net/handler"
)

var (
	pingReturn = &pb.Bool{Value: true}
)

type (
	Handler struct {
		ch.Handler
	}
)

func (p *Handler) Name() string {
	return "opsHandler"
}

// OnInit 注册remote函数
func (p *Handler) OnInit() {
	p.AddRemote("ping", p.ping)
}

// ping 请求center是否响应
func (p *Handler) ping() (*pb.Bool, int32) {
	return pingReturn, code.OK
}
