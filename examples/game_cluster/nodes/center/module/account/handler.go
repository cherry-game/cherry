package account

import (
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center/db"
	ch "github.com/cherry-game/cherry/net/handler"
)

type (
	Handler struct {
		ch.Handler
	}
)

func (p *Handler) Name() string {
	return "accountHandler"
}

// OnInit center为后端节点，不直接与客户端通信，所以了一些remote函数，供RPC调用
func (p *Handler) OnInit() {
	p.AddRemote("getDevAccount", p.getDevAccount)
	p.AddRemote("getUID", p.getUID)
}

// getDevAccount 根据帐号名获取开发者帐号表
func (p *Handler) getDevAccount(req *pb.DevRegister) (*pb.Int64, int32) {
	accountName := req.AccountName
	password := req.Password

	devAccount, _ := db.DevAccountWithName(accountName)
	if devAccount == nil || devAccount.Password != password {
		return nil, code.AccountAuthFail
	}

	return &pb.Int64{Value: devAccount.AccountId}, code.OK
}

// getUID 获取uid
func (p *Handler) getUID(req *pb.User) (*pb.Int64, int32) {
	uid, ok := db.BindUID(req.SdkId, req.Pid, req.OpenId)
	if uid == 0 || ok == false {
		return nil, code.AccountBindFail
	}

	return &pb.Int64{Value: uid}, code.OK
}
