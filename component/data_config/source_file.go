package cherryDataConfig

import (
	cerr "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/radovskyb/watcher"
	"io/ioutil"
	"time"
)

// SourceFile 本地读取数据配置文件
type SourceFile struct {
	monitorPath string
	watcher     *watcher.Watcher
	reloadTime  int64
	extName     string
	changeFn    ConfigChangeFn
}

func (f *SourceFile) Name() string {
	return "file"
}

func (f *SourceFile) Init(_ IDataConfig) {
	//read data_config->file node
	dataConfig := cherryProfile.GetConfig("data_config").GetConfig(f.Name())
	if dataConfig.LastError() != nil {
		cherryLogger.Warnf("[data_config]->[%s] node in `%s` file not found.", f.Name(), cherryProfile.FileName())
		return
	}

	f.extName = dataConfig.GetString("ext_name", ".json")
	f.reloadTime = dataConfig.GetInt64("reload_time", 3000)
	filePath := dataConfig.GetString("file_path", "data/")

	var err error
	f.monitorPath, err = cherryFile.JoinPath(cherryProfile.Dir(), filePath)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	// new watcher
	go f.newWatcher()
}

func (f *SourceFile) ReadBytes(configName string) (data []byte, error error) {
	if configName == "" {
		return nil, cerr.Error("configName is empty.")
	}

	fullPath, err := cherryFile.JoinPath(f.monitorPath, configName+f.extName)
	if err != nil {
		return nil, cerr.Errorf("file not found. err = %v", err)
	}

	if cherryFile.IsDir(fullPath) {
		return nil, cerr.Errorf("path is dir. fullPath = %s", fullPath)
	}

	data, err = ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, cerr.Errorf("read file err. err = %v, path = %s", err, fullPath)
	}

	if len(data) < 1 {
		return nil, cerr.Errorf("configName = %s data is err.", configName)
	}

	return data, nil
}

func (f *SourceFile) OnChange(fn ConfigChangeFn) {
	f.changeFn = fn
}

func (f *SourceFile) newWatcher() {
	f.watcher = watcher.New()
	f.watcher.SetMaxEvents(1)
	f.watcher.FilterOps(watcher.Write)

	if err := f.watcher.Add(f.monitorPath); err != nil {
		cherryLogger.Warn("new watcher error. path=%s, err=%v", f.monitorPath, err)
		return
	}

	//new goroutine
	go func() {
		for {
			select {
			case ev := <-f.watcher.Event:
				{
					if ev.IsDir() {
						return
					}

					configName := cherryFile.GetFileName(ev.FileInfo.Name(), true)
					cherryLogger.Infof("[name = %s] trigger file change.", configName)

					data, err := f.ReadBytes(configName)
					if err != nil {
						cherryLogger.Warn("[name = %s] read data error = %s", configName, err)
						return
					}

					if f.changeFn != nil {
						f.changeFn(configName, data)
					}
				}
			case err := <-f.watcher.Error:
				{
					cherryLogger.Error(err)
					return
				}
			case <-f.watcher.Closed:
				return
			}
		}
	}()

	if err := f.watcher.Start(time.Millisecond * time.Duration(f.reloadTime)); err != nil {
		cherryLogger.Warn(err)
	}
}

func (f *SourceFile) Stop() {
	if f.watcher != nil {
		err := f.watcher.Remove(f.monitorPath)
		if err != nil {
			cherryLogger.Warn(err)
		}
		cherryLogger.Infof("remove watcher [path = %s]", f.monitorPath)
		f.watcher.Closed <- struct{}{}
	}
}
