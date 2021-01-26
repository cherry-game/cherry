package cherryDataConfig

import (
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
)

type DataConfigComponent struct {
	cherryInterfaces.BaseComponent

	dataSource  IDataSource
	dataParse   IDataParse
	models      map[string]IConfigModel
	serviceList []IConfigService
}

func NewComponent() *DataConfigComponent {
	return &DataConfigComponent{}
}

func (d *DataConfigComponent) Init() {
	d.initModels()
}

func (d *DataConfigComponent) initModels() {
	for _, model := range d.models {
		bytes, err := d.dataSource.GetContent(model.FileName())
		if err != nil {
			cherryLogger.Error(err)
			continue
		}

		if len(bytes) < 1 {
			cherryLogger.Errorf("fileName=%s content is null", model.FileName())
			continue
		}

		var list []IConfigModel
		if err := d.dataParse.Parse(bytes, &list); err != nil {
			cherryLogger.Errorf("data parse error. error = %s", err)
			return
		}

		if len(list) < 1 {
			cherryLogger.Errorf("fileName=%s parse to list is empty.", model.FileName())
			continue
		}

		for _, m := range list {
			m.Init()
		}

	}
}

func (d *DataConfigComponent) GetFirst(index *IndexObject, params ...interface{}) interface{} {
	return nil
}

func (d *DataConfigComponent) GetList(tableName string) interface{} {
	return nil
}

func (d *DataConfigComponent) GetIndexList(index *IndexObject, params ...interface{}) interface{} {
	return nil
}

func (d *DataConfigComponent) Reload(fileName string, text []byte) error {
	return nil
}

func (d *DataConfigComponent) CheckFileName(fileNames string, text []byte) error {
	return nil
}

func (d *DataConfigComponent) RegisterModel(models ...IConfigModel) error {
	for _, model := range models {

		if len(model.FileName()) < 1 {
			return cherryUtils.ErrorFormat("model=%t, fileName() is nil", model)
		}

		if _, found := d.models[model.FileName()]; found {
			return cherryUtils.ErrorFormat("fileName=%s have duplicate.", model)
		}
		d.models[model.FileName()] = model
	}

	return nil
}

func (d *DataConfigComponent) RegisterService(service IConfigService) {
	d.serviceList = append(d.serviceList, service)
}
