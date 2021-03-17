package mocks

import (
	"github.com/cherry-game/cherry/extend/mapstructure"
	"github.com/cherry-game/cherry/extend/utils"
)

type (
	DropConfig struct {
		DropId    int    `json:"dropId"`
		ItemType  int    `json:"itemType"`
		ItemId    int    `json:"itemId"`
		Num       int    `json:"num"`
		DropType  int    `json:"dropType"`
		DropValue int    `json:"dropValue"`
		Desc      string `json:"desc"`
	}

	DropConfigs struct {
		list []*DropConfig
	}
)

func (d *DropConfigs) Name() string {
	return "dropConfig"
}

func (d *DropConfigs) Load(maps interface{}, reload bool) error {
	list, ok := maps.([]interface{})
	if ok == false {
		return cherryUtils.Error("maps convert to []interface{} error.")
	}

	if reload {
		d.list = d.list[0:0]
	}

	for _, data := range list {
		var item DropConfig
		err := cherryMapStructure.Decode(data, &item)
		if err == nil {
			d.list = append(d.list, &item)
		}
	}

	return nil
}

func (d *DropConfigs) Get(dropId int) *DropConfig {
	for _, config := range d.list {
		if config.DropId == dropId {
			return config
		}
	}
	return nil
}
