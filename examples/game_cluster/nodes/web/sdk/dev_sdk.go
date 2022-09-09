package sdk

import (
	cherryGin "github.com/cherry-game/cherry/components/gin"
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	rpcCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/rpc/center"
	cherryString "github.com/cherry-game/cherry/extend/string"
)

type devSdk struct {
}

func (devSdk) SdkId() int32 {
	return DevMode
}

func (devSdk) Login(_ *data.SdkRow, params Params, callback Callback) {
	accountName, _ := params.GetString("account")
	password, _ := params.GetString("password")

	if accountName == "" || password == "" {
		err := cherryError.Errorf("account or password params is empty.")
		callback(code.LoginError, nil, err)
		return
	}

	accountId := rpcCenter.GetDevAccount(accountName, password)
	if accountId < 1 {
		callback(code.LoginError, nil)
		return
	}

	callback(code.OK, map[string]string{
		"open_id": cherryString.ToString(accountId),
	})
}

func (devSdk) PayCallback(_ *data.SdkRow, _ *cherryGin.Context) {
}
