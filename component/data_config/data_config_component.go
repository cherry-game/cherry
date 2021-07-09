package cherryDataConfig

import (
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"sync"
)

type Component struct {
	sync.RWMutex
	cherryFacade.Component
	dataSource IDataSource
	parser     IDataParser
	configs    []IConfig
	configMaps map[string]IConfig
}

func NewComponent() *Component {
	return &Component{
		configMaps: make(map[string]IConfig),
	}
}

// Name unique components name
func (d *Component) Name() string {
	return cherryConst.DataConfigComponent
}

func (d *Component) Init() {
	// read data_config node in profile-{env}.json
	configNode := cherryProfile.GetConfig("data_config")
	if configNode.LastError() != nil {
		panic(fmt.Sprintf("`data_config` node in `%s` file not found.", cherryProfile.FileName()))
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

		// read register IConfig
		for _, cfg := range d.configs {
			data, found := d.GetBytes(cfg.Name())
			if !found {
				cherryLogger.Warnf("configName = %s not found.", cfg.Name())
				continue
			}
			d.initConfig(cfg, data, false)
		}

		// on change process
		d.dataSource.OnChange(func(configName string, data []byte) {
			configFile := d.GetIConfigFile(configName)
			if configFile != nil {
				d.initConfig(configFile, data, true)
			}
		})

	}, func(errString string) {
		cherryLogger.Warn(errString)
	})
}

func (d *Component) initConfig(cfg IConfig, data []byte, reload bool) {
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

	d.configMaps[cfg.Name()] = cfg
}

func (d *Component) OnStop() {
	if d.dataSource != nil {
		d.dataSource.Stop()
	}
}

func (d *Component) Register(configs ...IConfig) {
	if len(configs) < 1 {
		cherryLogger.Warnf("IConfig size is less than 1.")
		return
	}

	for _, cfg := range configs {
		if cfg != nil {
			d.configs = append(d.configs, cfg)
		}
	}
}

func (d *Component) GetIConfigFile(name string) IConfig {
	for _, file := range d.configs {
		if file.Name() == name {
			return file
		}
	}
	return nil
}

func (d *Component) GetBytes(configName string) (data []byte, found bool) {
	data, err := d.dataSource.ReadBytes(configName)
	if err != nil {
		cherryLogger.Warn(err)
		return nil, false
	}

	return data, true
}

func (d *Component) GetParser() IDataParser {
	return d.parser
}

func (d *Component) GetDataSource() IDataSource {
	return d.dataSource
}
