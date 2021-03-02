package mocks

type DropConfig struct {
	DropId    int `json:"dropId"`
	ItemType  int `json:"itemType"`
	ItemId    int `json:"itemId"`
	Num       int `json:"num"`
	DropType  int `json:"dropType"`
	DropValue int `json:"dropValue"`
}

func (d *DropConfig) Name() string {
	return "dropConfig"
}

func (d *DropConfig) Init() {

}

func (d *DropConfig) Reload() {

}
