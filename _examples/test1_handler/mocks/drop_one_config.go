package mocks

type DropOneConfig struct {
	DropId    int `json:"dropId"`
	ItemType  int `json:"itemType"`
	ItemId    int `json:"itemId"`
	Num       int `json:"num"`
	DropType  int `json:"dropType"`
	DropValue int `json:"dropValue"`
}

func (d *DropOneConfig) Name() string {
	return "dropOneConfig"
}

func (d *DropOneConfig) Init() {

}

func (d *DropOneConfig) Reload() {

}
