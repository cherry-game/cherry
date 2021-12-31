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
	configNode := cherryProfile.Config().Get("data_config")
	if configNode.LastError() != nil {
		panic(fmt.Sprintf("`data_config` node in `%s` file not found.", cherryProfile.FileName()))
	}

	// get data source
	sourceName := configNode.Get("data_source").ToString()
	d.dataSource = GetDataSource(sourceName)
	if d.dataSource == nil {
		panic(fmt.Sprintf("[sourceName = %s] data source not found.", sourceName))
	}

	// get parser
	parserName := configNode.Get("parser").ToString()
	d.parser = GetParser(parserName)
	if d.parser == nil {
		panic(fmt.Sprintf("[parserName = %s] parser not found.", parserName))
	}

	cherryUtils.Try(func() {
		d.dataSource.Init(d)

		// init
		for _, cfg := range d.configs {
			cfg.Init()
		}

		// read register IConfig
		for _, cfg := range d.configs {
			data, found := d.GetBytes(cfg.Name())
			if !found {
				cherryLogger.Warnf("[config = %s] load data fail.", cfg.Name())
				continue
			}

			cherryUtils.Try(func() {
				d.onLoadConfig(cfg, data, false)
			}, func(errString string) {
				cherryLogger.Errorf("[config = %s] init config error. [error = %s]", cfg.Name(), errString)
			})
		}

		// on after load
		for _, cfg := range d.configs {
			cfg.OnAfterLoad(false)
		}

		// on change process
		d.dataSource.OnChange(func(configName string, data []byte) {
			configFile := d.GetIConfigFile(configName)
			if configFile != nil {
				d.onLoadConfig(configFile, data, true)
				configFile.OnAfterLoad(true)
			}
		})

	}, func(errString string) {
		cherryLogger.Error(errString)
	})
}

func (d *Component) onLoadConfig(cfg IConfig, data []byte, reload bool) {
	d.Lock()
	defer d.Unlock()

	var parseObject interface{}
	err := d.parser.Unmarshal(data, &parseObject)
	if err != nil {
		cherryLogger.Warnf("[config = %s] unmarshal error = %v", err, cfg.Name())
		return
	}

	// load data
	size, err := cfg.OnLoad(parseObject, reload)
	if err != nil {
		cherryLogger.Warnf("[config = %s] execute Load() error = %s", cfg.Name(), err)
		return
	}

	cherryLogger.Infof("[config = %s] loaded. [size = %d]", cfg.Name(), size)

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
