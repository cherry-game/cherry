package data

import (
	"github.com/cherry-game/cherry"
	cherryDataConfig "github.com/cherry-game/cherry/components/data-config"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/types"
	cherryMapStructure "github.com/cherry-game/cherry/extend/mapstructure"
)

var (
	AreaConfig       = &areaConfig{}
	AreaGroupConfig  = &areaGroupConfig{}
	AreaServerConfig = &areaServerConfig{}
	SdkConfig        = &sdkConfig{}
	CodeConfig       = &codeConfig{}
)

func RegisterComponent() {
	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(
		AreaConfig,
		AreaGroupConfig,
		AreaServerConfig,
		SdkConfig,
		CodeConfig,
	)

	cherry.RegisterComponent(dataConfig)
}

func DecodeData(input interface{}, output interface{}) error {
	return cherryMapStructure.HookDecode(
		input,
		output,
		"json",
		types.GetDecodeHooks(),
	)
}
