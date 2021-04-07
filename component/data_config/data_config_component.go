package cherryDataConfig

import (
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"sync"
)

type DataConfigComponent struct {
	sync.RWMutex
	cherryInterfaces.BaseComponent
	dataSource     IDataSource
	parser         IDataParser
	configFiles    []IConfigFile
	configDataMaps map[string]IConfigFile
}

func NewComponent() *DataConfigComponent {
	return &DataConfigComponent{
		configDataMaps: make(map[string]IConfigFile),
	}
}

// Name unique components name
func (d *DataConfigComponent) Name() string {
	return cherryConst.DataConfigComponent
}

func (d *DataConfigComponent) Init() {
	// read data_config node in profile-{env}.json
	configNode := cherryProfile.GetConfig("data_config")
	if configNode.LastError() != nil {
		panic(fmt.Sprintf("not found `data_config` node in `%s` file.", cherryProfile.FileName()))
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

		// read register IConfigFile
		for _, cfg := range d.configFiles {
			data, found := d.GetBytes(cfg.Name())
			if !found {
				cherryLogger.Warnf("configName = %s not found.", cfg.Name())
				continue
			}
			d.initConfigFile(cfg, data, false)
		}

		// on change process
		d.dataSource.OnChange(func(configName string, data []byte) {
			configFile := d.GetIConfigFile(configName)
			if configFile != nil {
				d.initConfigFile(configFile, data, true)
			}
		})

	}, func(errString string) {
		cherryLogger.Warn(errString)
	})
}

func (d *DataConfigComponent) initConfigFile(cfg IConfigFile, data []byte, reload bool) {
	d.Lock()
	defer d.Unlock()

	var parseObject interface{}
	err := d.parser.Unmarshal(data, &parseObject)
	if err != nil {
		cherryLogger.Warnf("unmarshal data error = %v, configName = %s", err, cfg.Name())
		return
	}

	// load data
	err = cfg.Load(parseObject, reload)
	if err != nil {
		cherryLogger.Warnf("read name = %s on init error = %s", cfg.Name(), err)
	}

	d.configDataMaps[cfg.Name()] = cfg
}

func (d *DataConfigComponent) Stop() {
	if d.dataSource != nil {
		d.dataSource.Stop()
	}
}

func (d *DataConfigComponent) Register(configFiles ...IConfigFile) {
	if len(configFiles) < 1 {
		cherryLogger.Warnf("IConfigFile size is less than 1.")
		return
	}

	for _, cfg := range configFiles {
		if cfg != nil {
			d.configFiles = append(d.configFiles, cfg)
		}
	}
}

func (d *DataConfigComponent) GetIConfigFile(name string) IConfigFile {
	for _, file := range d.configFiles {
		if file.Name() == name {
			return file
		}
	}
	return nil
}

func (d *DataConfigComponent) GetBytes(configName string) (data []byte, found bool) {
	data, err := d.dataSource.ReadBytes(configName)
	if err != nil {
		cherryLogger.Warn(err)
		return nil, false
	}

	return data, true
}

func (d *DataConfigComponent) GetParser() IDataParser {
	return d.parser
}

func (d *DataConfigComponent) GetDataSource() IDataSource {
	return d.dataSource
}
