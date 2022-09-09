package data

import (
	cherryError "github.com/cherry-game/cherry/error"
	cherryLogger "github.com/cherry-game/cherry/logger"
)

type (
	AreaGroupRow struct {
		PID        int32   `json:"pid"`        // 渠道id
		AreaIdList []int32 `json:"areaIdList"` // 归属的游戏区id列表
	}

	// 游戏区分组
	areaGroupConfig struct {
		maps map[int32]AreaGroupRow
	}
)

// Name 根据名称读取 ./config/data/areaGroupConfig.json文件
func (p *areaGroupConfig) Name() string {
	return "areaGroupConfig"
}

func (p *areaGroupConfig) Init() {
	p.maps = make(map[int32]AreaGroupRow)
}

func (p *areaGroupConfig) OnLoad(maps interface{}, _ bool) (int, error) {
	list, ok := maps.([]interface{})
	if !ok {
		return 0, cherryError.Error("maps convert to []interface{} error.")
	}

	loadMaps := make(map[int32]AreaGroupRow)
	for index, data := range list {
		loadConfig := AreaGroupRow{}
		err := DecodeData(data, &loadConfig)
		if err != nil {
			cherryLogger.Warnf("decode error. [row = %d, %v], err = %s", index+1, loadConfig, err)
			continue
		}

		loadMaps[loadConfig.PID] = loadConfig
	}

	p.maps = loadMaps

	return len(list), nil
}

func (p *areaGroupConfig) OnAfterLoad(_ bool) {
}

func (p *areaGroupConfig) Get(pk int32) (*AreaGroupRow, bool) {
	i, found := p.maps[pk]
	return &i, found
}

func (p *areaGroupConfig) Contain(pk int32) bool {
	_, found := p.Get(pk)
	return found
}
