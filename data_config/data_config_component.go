package cherryDataConfig

import (
	"github.com/phantacix/cherry/interfaces"
	"github.com/phantacix/cherry/logger"
	"github.com/phantacix/cherry/utils"
)

type DefaultComponent struct {
	cherryInterfaces.BaseComponent

	dataSource  IDataSource
	dataParse   IDataParse
	models      map[string]IConfigModel
	serviceList []IConfigService
}

func NewComponent() *DefaultComponent {
	return &DefaultComponent{}
}

func (d *DefaultComponent) Init() {
	d.initModels()
}

func (d *DefaultComponent) initModels() {
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
		d.dataParse.Parse(bytes, &list)

		if len(list) < 1 {
			cherryLogger.Errorf("fileName=%s parse to list is empty.", model.FileName())
			continue
		}

		for _, m := range list {
			m.Init()
		}

	}
}

func (d *DefaultComponent) GetFirst(index *IndexObject, params ...interface{}) interface{} {
	return nil
}

func (d *DefaultComponent) GetList(tableName string) interface{} {
	return nil
}

func (d *DefaultComponent) GetIndexList(index *IndexObject, params ...interface{}) interface{} {
	return nil
}

func (d *DefaultComponent) Reload(fileName string, text []byte) error {
	return nil
}

func (d *DefaultComponent) CheckFileName(fileNames string, text []byte) error {
	return nil
}

func (d *DefaultComponent) RegisterModel(models ...IConfigModel) error {
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

func (d *DefaultComponent) RegisterService(service IConfigService) {
	d.serviceList = append(d.serviceList, service)
}
