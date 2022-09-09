package account

import (
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center/db"
	ch "github.com/cherry-game/cherry/net/handler"
	"strings"
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
	p.AddRemote("registerDevAccount", p.registerDevAccount)
	p.AddRemote("getDevAccount", p.getDevAccount)
	p.AddRemote("getUID", p.getUID)
}

// registerDevAccount 注册开发者帐号
func (p *Handler) registerDevAccount(req *pb.DevRegister) int32 {
	accountName := req.AccountName
	password := req.Password

	if strings.TrimSpace(accountName) == "" || strings.TrimSpace(password) == "" {
		return code.LoginError
	}

	if len(accountName) < 3 || len(accountName) > 18 {
		return code.LoginError
	}

	if len(password) < 3 || len(password) > 18 {
		return code.LoginError
	}

	return db.DevAccountRegister(accountName, password, req.Ip)
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
