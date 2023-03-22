package rpcCenter

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

// route = 节点类型.节点handler.remote函数

const (
	centerType = "center"
)

const (
	opsActor     = ".ops"
	accountActor = ".account"
)

const (
	ping               = "ping"
	registerDevAccount = "registerDevAccount"
	getDevAccount      = "getDevAccount"
	getUID             = "getUID"
)

// Ping 访问center节点，确认center已启动
func Ping(app cfacade.IApplication) bool {
	nodeId := GetCenterNodeID(app)
	if nodeId == "" {
		return false
	}

	rsp := &pb.Bool{}
	targetPath := nodeId + opsActor
	errCode := app.ActorSystem().CallWait("", targetPath, ping, nil, rsp)
	if code.IsFail(errCode) {
		return false
	}

	return rsp.Value
}

// RegisterDevAccount 注册帐号
func RegisterDevAccount(app cfacade.IApplication, accountName, password, ip string) int32 {
	req := &pb.DevRegister{
		AccountName: accountName,
		Password:    password,
		Ip:          ip,
	}

	targetPath := GetTargetPath(app, accountActor)
	rsp := &pb.Int32{}
	errCode := app.ActorSystem().CallWait("", targetPath, registerDevAccount, req, rsp)
	if code.IsFail(errCode) {
		clog.Warnf("[RegisterDevAccount] accountName = %s, errCode = %v", accountName, errCode)
		return errCode
	}

	return rsp.Value
}

// GetDevAccount 获取帐号信息
func GetDevAccount(app cfacade.IApplication, accountName, password string) int64 {
	req := &pb.DevRegister{
		AccountName: accountName,
		Password:    password,
	}

	targetPath := GetTargetPath(app, accountActor)
	rsp := &pb.Int64{}
	errCode := app.ActorSystem().CallWait("", targetPath, getDevAccount, req, rsp)
	if code.IsFail(errCode) {
		clog.Warnf("[GetDevAccount] accountName = %s, errCode = %v", accountName, errCode)
		return 0
	}

	return rsp.Value
}

// GetUID 获取帐号UID
func GetUID(app cfacade.IApplication, sdkId, pid int32, openId string) (cfacade.UID, int32) {
	req := &pb.User{
		SdkId:  sdkId,
		Pid:    pid,
		OpenId: openId,
	}

	targetPath := GetTargetPath(app, accountActor)
	rsp := &pb.Int64{}
	errCode := app.ActorSystem().CallWait("", targetPath, getUID, req, rsp)
	if code.IsFail(errCode) {
		clog.Warnf("[GetUID] errCode = %v", errCode)
		return 0, errCode
	}

	return rsp.Value, code.OK
}

func GetCenterNodeID(app cfacade.IApplication) string {
	list := app.Discovery().ListByType(centerType)
	if len(list) > 0 {
		return list[0].GetNodeId()
	}
	return ""
}

func GetTargetPath(app cfacade.IApplication, actorID string) string {
	nodeId := GetCenterNodeID(app)
	return nodeId + actorID
}
