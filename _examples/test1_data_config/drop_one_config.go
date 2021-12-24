package main

import "github.com/cherry-game/cherry/extend/mapstructure"

type DropOneConfig struct {
	DropId    int    `json:"dropId"`
	ItemType  int    `json:"itemType"`
	ItemId    int    `json:"itemId"`
	Num       int    `json:"num"`
	DropType  int    `json:"dropType"`
	DropValue int    `json:"dropValue"`
	Desc      string `json:"desc"`
}

func (d *DropOneConfig) Name() string {
	return "dropOneConfig"
}

func (d *DropOneConfig) Init() {

}

func (d *DropOneConfig) Load(maps interface{}, _ bool) (int, error) {
	err := cherryMapStructure.Decode(maps, d)
	return 0, err
}
