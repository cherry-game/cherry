package mocks

type DropConfig struct {
	dropId    int `json:dropId`
	itemType  int `json:itemType`
	itemId    int `json:itemId`
	num       int `json:num`
	dropType  int `json:dropType`
	dropValue int `json:dropValue`
}

func (d *DropConfig) Name() string {
	return "dropConfig"
}

func (d *DropConfig) Init() {

}

func (d *DropConfig) Reload() {

}
