package sdk

import (
	cherryGin "github.com/cherry-game/cherry/components/gin"
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	cherryString "github.com/cherry-game/cherry/extend/string"
)

// sdk平台类型
const (
	DevMode  int32 = 1 // 开发模式，注册开发帐号登陆(开发时使用)
	QuickSDK int32 = 2 // quick sdk
)

var (
	invokeMaps = make(map[int32]Invoke)
)

type (
	Invoke interface {
		SdkId() int32                                                // sdk id
		Login(config *data.SdkRow, params Params, callback Callback) // Login 登录验证接口
		PayCallback(config *data.SdkRow, c *cherryGin.Context)       // PayCallback 支付回调接口
	}

	Params map[string]string

	Callback func(code int32, result Params, error ...error)
)

func (p Params) GetInt(key string, defaultValue ...int) int {
	defVal := 0
	if len(defaultValue) > 0 {
		defVal = defaultValue[0]
	}

	val, found := p[key]
	if !found {
		return defVal
	}

	intVal, ok := cherryString.ToInt(val)
	if ok {
		return intVal
	}

	return defVal
}

func (p Params) GetString(key string) (string, bool) {
	v, ok := p[key]
	return v, ok
}

func register(invoke Invoke) {
	invokeMaps[invoke.SdkId()] = invoke
}

func GetInvoke(sdkId int32) (invoke Invoke, error error) {
	invoke, found := invokeMaps[sdkId]
	if found == false {
		return nil, cherryError.Errorf("[sdkId = %d] not found.", sdkId)
	}

	return invoke, nil
}

func Init() {
	register(devSdk{})
	register(quickSdk{})
}
