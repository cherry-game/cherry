package main

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/extend/mapstructure"
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

func (d *DropConfigs) Init() {

}

func (d *DropConfigs) OnLoad(maps interface{}, reload bool) (int, error) {
	list, ok := maps.([]interface{})
	if ok == false {
		return 0, cherryError.Error("maps convert to []interface{} error.")
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

	return len(list), nil
}

func (d *DropConfigs) OnAfterLoad(reload bool) {
}

func (d *DropConfigs) Get(dropId int) *DropConfig {
	for _, config := range d.list {
		if config.DropId == dropId {
			return config
		}
	}
	return nil
}

func (d *DropConfigs) GetItemTypeList(itemType int) []*DropConfig {
	var c []*DropConfig

	linq.From(d.list).Where(func(i interface{}) bool {
		return i.(*DropConfig).ItemType == itemType
	}).ToSlice(&c)

	return c
}
