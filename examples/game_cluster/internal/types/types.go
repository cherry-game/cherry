package types

import (
	cherryMapStructure "github.com/cherry-game/cherry/extend/mapstructure"
	"reflect"
)

type (
	HookType interface {
		Type() reflect.Type
		Hook() cherryMapStructure.DecodeHookFuncType
	}
)

var (
	funcTypes []cherryMapStructure.DecodeHookFuncType
)

func init() {
	// 需要通过json解析数据的类型，注册到此
	register(&I32I32{})
	register(&I32I64Map{})
}

func register(t HookType) {
	funcTypes = append(funcTypes, t.Hook())
}

func GetDecodeHooks() []cherryMapStructure.DecodeHookFuncType {
	return funcTypes
}
