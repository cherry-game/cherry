package cherryDataConfig

import (
	"github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/radovskyb/watcher"
	"io/ioutil"
	"time"
)

// FileSource 本地读取数据配置文件
type FileSource struct {
	dataConfig   IDataConfig
	monitorPath  string
	watcher      *watcher.Watcher
	reloadTime   int64
	extName      string
	changeFileFn ChangeFileFn
}

func (f *FileSource) Name() string {
	return "file"
}

func (f *FileSource) Init(dataConfig IDataConfig) {
	f.dataConfig = dataConfig

	//read data_config->file node
	fileNode := cherryProfile.Config("data_config", "file")
	if fileNode == nil {
		cherryLogger.Warnf("`data_config` node not found in `%s` file.", cherryProfile.FileName())
		return
	}

	filePath := fileNode.Get("file_path").ToString()
	if filePath == "" {
		filePath = "data_config/" //default value
	}

	f.extName = fileNode.Get("ext_name").ToString()
	if f.extName == "" {
		f.extName = ".json" // default value
	}

	var err error
	f.monitorPath, err = cherryFile.JoinPath(cherryProfile.Dir(), filePath)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	f.reloadTime = fileNode.Get("reload_time").ToInt64()
	if f.reloadTime < 1 {
		f.reloadTime = 2000 //default value
	}

	// new watcher
	go f.newWatcher()
}

func (f *FileSource) ReadData(configName string) (data []byte, error error) {
	if configName == "" {
		return nil, cherryUtils.Error("configName is empty.")
	}

	fullPath, error := cherryFile.JoinPath(f.monitorPath, configName+f.extName)
	if error != nil {
		return nil, cherryUtils.Errorf("file not found. err = %v, fullPath = %s", error, fullPath)
	}

	if cherryFile.IsDir(fullPath) {
		return nil, cherryUtils.Errorf("path is dir. fullPath = %s", error, fullPath)
	}

	data, error = ioutil.ReadFile(fullPath)
	if error != nil {
		return nil, cherryUtils.Errorf("read file err. err = %v path = %s", error, fullPath)
	}

	if len(data) < 1 {
		return nil, cherryUtils.Error("configName data is error.")
	}

	return data, nil
}

func (f *FileSource) OnChange(changeFileFn ChangeFileFn) {
	f.changeFileFn = changeFileFn
}

func (f *FileSource) newWatcher() {
	f.watcher = watcher.New()
	f.watcher.SetMaxEvents(1)
	f.watcher.FilterOps(watcher.Write)

	err := f.watcher.Add(f.monitorPath)
	if err != nil {
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

					data, err := f.ReadData(configName)
					if err != nil {
						cherryLogger.Error(err)
						return
					}

					f.changeFileFn(configName, data)

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

func (f *FileSource) Stop() {
	if f.watcher != nil {
		err := f.watcher.Remove(f.monitorPath)
		if err != nil {
			cherryLogger.Warn(err)
		}
		cherryLogger.Infof("remove watcher [path = %s]", f.monitorPath)
		f.watcher.Closed <- struct{}{}
	}
}
