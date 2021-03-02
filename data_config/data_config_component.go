package cherryDataConfig

import (
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/goinggo/mapstructure"
	"reflect"
	"sync"
)

type DataConfigComponent struct {
	sync.Mutex
	cherryInterfaces.BaseComponent
	configNames  []string
	registerMaps map[string]interface{}
	dataSource   IDataSource
	parser       IParser
}

func NewComponent() *DataConfigComponent {
	return &DataConfigComponent{
		registerMaps: make(map[string]interface{}),
		configNames:  []string{},
	}
}

// Name unique components name
func (d *DataConfigComponent) Name() string {
	return cherryConst.DataConfigComponent
}

func (d *DataConfigComponent) Init() {
	// read data_config node in profile-{env}.json
	configNode := cherryProfile.Config("data_config")
	if configNode.LastError() != nil {
		panic(fmt.Sprintf("not found `data_config` node in `%s` file.", cherryProfile.FilePath()))
	}

	// get data source
	sourceName := configNode.Get("data_source").ToString()
	d.dataSource = GetDataSource(sourceName)
	if d.dataSource == nil {
		panic(fmt.Sprintf("data source not found. sourceName = %s", sourceName))
	}

	// get parser
	parserName := configNode.Get("parser").ToString()
	d.parser = GetParser(parserName)
	if d.parser == nil {
		panic(fmt.Sprintf("parser not found. parserName = %s", parserName))
	}

	cherryUtils.Try(func() {
		d.dataSource.Init(d)

		for _, name := range d.configNames {
			data, found := d.GetBytes(name)
			if !found {
				cherryLogger.Warnf("configName = %s not found.", name)
				continue
			}

			var val interface{}
			err := d.parser.Unmarshal(data, &val)
			if err != nil {
				cherryLogger.Warnf("unmarshal data error=%v, configName=%s", err, name)
				continue
			}

			d.registerMaps[name] = val
		}
	}, func(errString string) {
		cherryLogger.Warn(errString)
	})
}

func (d *DataConfigComponent) Stop() {
	if d.dataSource != nil {
		d.dataSource.Stop()
	}
}

func (d *DataConfigComponent) Register(configNames ...string) {
	if len(configNames) < 1 {
		return
	}

	for _, name := range configNames {
		if name != "" {
			d.configNames = append(d.configNames, name)
		}
	}
}

func (d *DataConfigComponent) GetBytes(configName string) (data []byte, found bool) {
	data, err := d.dataSource.ReadData(configName)
	if err != nil {
		cherryLogger.Warn(err)
		return nil, false
	}

	return data, true
}

func (d *DataConfigComponent) Get(configName string, val interface{}) {
	typ := reflect.TypeOf(val)
	if typ.Kind() != reflect.Ptr {
		cherryLogger.Warnf("val must ptr type. configName={}", configName)
		return
	}

	result, found := d.registerMaps[configName]
	if found {
		if err := mapstructure.Decode(result, val); err != nil {
			cherryLogger.Warn(err)
		}
	}
}

func (d *DataConfigComponent) GetParser() IParser {
	return d.parser
}

func (d *DataConfigComponent) GetDataSource() IDataSource {
	return d.dataSource
}
