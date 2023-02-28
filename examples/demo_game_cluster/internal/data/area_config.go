package data

import (
	cherryError "github.com/cherry-game/cherry/error"
	cherryLogger "github.com/cherry-game/cherry/logger"
)

type (
	AreaRow struct {
		AreaId   int32  `json:"areaId"`   // 游戏区id
		AreaName string `json:"areaName"` // 游戏区名称
		Gate     string `json:"gate"`     // 游戏区对应的网关地址
	}

	// 游戏区
	areaConfig struct {
		maps map[int32]*AreaRow
	}
)

// Name 根据名称读取 ./config/data/areaConfig.json文件
func (p *areaConfig) Name() string {
	return "areaConfig"
}

func (p *areaConfig) Init() {
	p.maps = make(map[int32]*AreaRow)
}

func (p *areaConfig) OnLoad(maps interface{}, _ bool) (int, error) {
	list, ok := maps.([]interface{})
	if !ok {
		return 0, cherryError.Error("maps convert to []interface{} error.")
	}

	loadMaps := make(map[int32]*AreaRow)
	for index, data := range list {
		loadConfig := &AreaRow{}
		err := DecodeData(data, loadConfig)
		if err != nil {
			cherryLogger.Warnf("decode error. [row = %d, %v], err = %s", index+1, loadConfig, err)
			continue
		}

		loadMaps[loadConfig.AreaId] = loadConfig
	}

	p.maps = loadMaps

	return len(list), nil
}

func (p *areaConfig) OnAfterLoad(_ bool) {
}

func (p *areaConfig) Get(pk int32) (*AreaRow, bool) {
	i, found := p.maps[pk]
	return i, found
}

func (p *areaConfig) Contain(pk int32) bool {
	_, found := p.Get(pk)
	return found
}
