package cherryDataConfig

import (
	"fmt"
	cherryUtils "github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"sync"
)

type DataConfigComponent struct {
	cherryInterfaces.BaseComponent
	sync.Mutex
	register      []IConfigFile
	configFiles   map[string]interface{}
	source        IDataSource
	parser        IParser
	parserExtName string
}

func NewComponent() *DataConfigComponent {
	return &DataConfigComponent{}
}

//Name unique components name
func (d *DataConfigComponent) Name() string {
	return "data_config_component"
}

func (d *DataConfigComponent) Init() {
	d.configFiles = make(map[string]interface{})

	// read data_config node in profile-x.json
	configNode := cherryProfile.Config().Get("data_config")
	if configNode.LastError() != nil {
		panic(fmt.Sprintf("not found `data_config` node in `%s` file.", cherryProfile.FilePath()))
	}

	// get data source
	sourceName := configNode.Get("data_source").ToString()
	d.source = GetDataSource(sourceName)
	if d.source == nil {
		panic(fmt.Sprintf("data source not found. sourceName = %s", sourceName))
	}

	// get file parser
	parserName := configNode.Get("parser").ToString()
	d.parser = GetParser(parserName)
	if d.parser == nil {
		panic(fmt.Sprintf("parser not found. sourceName = %s", parserName))
	}

	cherryUtils.Try(func() {
		d.source.Init(d)
	}, func(errString string) {
		cherryLogger.Warn(errString)
	})
}

func (d *DataConfigComponent) Stop() {
	if d.source != nil {
		d.source.Stop()
	}
}

func (d *DataConfigComponent) Register(file IConfigFile) {
	d.register = append(d.register, file)
}

func (d *DataConfigComponent) GetFiles() []IConfigFile {
	return d.register
}

func (d *DataConfigComponent) Get(fileName string) interface{} {
	return d.configFiles[fileName]
}

func (d *DataConfigComponent) Load(fileName string, data []byte) {
	cherryUtils.Try(func() {

		var v interface{}
		err := d.parser.Unmarshal(data, &v)
		if err != nil {
			cherryLogger.Warn(err)
			return
		}

		defer d.Unlock()
		d.Lock()

		d.configFiles[fileName] = &v

	}, func(errString string) {
		cherryLogger.Warn(errString)
	})
}
