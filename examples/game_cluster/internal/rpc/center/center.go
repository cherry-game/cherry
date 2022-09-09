package rpcCenter

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	cherryFacade "github.com/cherry-game/cherry/facade"
)

// route = 节点类型.节点handler.remote函数

const (
	opsHandler     = "center.opsHandler."
	accountHandler = "center.accountHandler."
)

// route
const (
	ping               = opsHandler + "ping"
	registerDevAccount = accountHandler + "registerDevAccount"
	getDevAccount      = accountHandler + "getDevAccount"
	getUID             = accountHandler + "getUID"
)

// Ping center节点是否响应
func Ping() bool {
	rsp := &pb.Bool{}
	statusCode := cherry.RequestRemoteByRoute(ping, nil, rsp)
	if code.IsFail(statusCode) {
		return false
	}

	return rsp.Value
}

func RegisterDevAccount(accountName, password, ip string) int32 {
	req := &pb.DevRegister{
		AccountName: accountName,
		Password:    password,
		Ip:          ip,
	}

	return cherry.RequestRemoteByRoute(registerDevAccount, req, nil)
}

func GetDevAccount(accountName, password string) int64 {
	req := &pb.DevRegister{
		AccountName: accountName,
		Password:    password,
	}

	rsp := &pb.Int64{}
	statusCode := cherry.RequestRemoteByRoute(getDevAccount, req, rsp)
	if code.IsFail(statusCode) {
		return 0
	}

	return rsp.Value
}

func GetUID(sdkId, pid int32, openId string) (cherryFacade.UID, int32) {
	req := &pb.User{
		SdkId:  sdkId,
		Pid:    pid,
		OpenId: openId,
	}

	rsp := &pb.Int64{}
	statusCode := cherry.RequestRemoteByRoute(getUID, req, rsp)

	return rsp.Value, statusCode
}
