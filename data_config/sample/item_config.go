package cherryDataConfigSample

import (
	"github.com/phantacix/cherry/data_config"
)

var (
	ITEM_CONFIG = "itemConfig"
)

type ItemConfig struct {
	Id   int
	Name string
}

func (i *ItemConfig) FileName() string {
	return ITEM_CONFIG
}

func (i *ItemConfig) Init() {

}

func (i *ItemConfig) ReadItem(item cherryDataConfig.IConfigModel) interface{} {
	return item.(*ItemConfig)
}

func (i ItemConfig) AddIndex() []*cherryDataConfig.IndexObject {
	var idx []*cherryDataConfig.IndexObject
	idx = append(idx, cherryDataConfig.NewIndex("itemConfig", "id"))
	return idx
}
