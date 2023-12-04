package cherryDataConfig

import (
	"os"
	"regexp"
	"time"

	cerr "github.com/cherry-game/cherry/error"
	cfile "github.com/cherry-game/cherry/extend/file"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/radovskyb/watcher"
)

type (
	// SourceFile 本地读取数据配置文件
	SourceFile struct {
		fileConfig
		watcher     *watcher.Watcher
		changeFn    ConfigChangeFn
		monitorPath string
	}

	fileConfig struct {
		FilePath   string `json:"file_path"`   // 配置文件路径
		ExtName    string `json:"ext_name"`    // 文件扩展名
		ReloadTime int64  `json:"reload_time"` // 定时重载扫描(毫秒)
	}
)

func (f *SourceFile) Name() string {
	return "file"
}

func (f *SourceFile) Init(_ IDataConfig) {
	err := f.unmarshalFileConfig()
	if err != nil {
		clog.Panicf("Unmarshal fileConfig fail. err = %v", err)
		return
	}

	f.watcher = watcher.New()
	f.watcher.FilterOps(watcher.Write)
	var regexpFilter *regexp.Regexp
	regexpFilter, err = regexp.Compile(`.*\` + f.ExtName + `$`)
	if err != nil {
		clog.Panicf("AddFilterHook extName fail. err = %v", err)
		return
	}
	f.watcher.AddFilterHook(watcher.RegexFilterHook(regexpFilter, false))

	f.monitorPath, err = cfile.JoinPath(cprofile.Path(), f.FilePath)
	if err != nil {
		clog.Panicf("[name = %s] join path fail. err = %v.", f.Name(), err)
		return
	}

	err = f.watcher.Add(f.monitorPath)
	if err != nil {
		clog.Panicf("New watcher error. path=%s, err = %v", f.monitorPath, err)
		return
	}

	// new watcher
	go f.newWatcher()
}

func (f *SourceFile) ReadBytes(configName string) ([]byte, error) {
	if configName == "" {
		return nil, cerr.Error("Config name is empty.")
	}

	fullPath, err := cfile.JoinPath(f.monitorPath, configName+f.ExtName)
	if err != nil {
		return nil, cerr.Errorf("Config file not found. err = %v", err)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, cerr.Errorf("Read file error. [path = %s, err = %v]", fullPath, err)
	}

	if len(data) < 1 {
		return nil, cerr.Errorf("Data is empty. [configName = %s]", configName)
	}

	return data, err
}

func (f *SourceFile) OnChange(fn ConfigChangeFn) {
	f.changeFn = fn
}

func (f *SourceFile) newWatcher() {
	go func() {
		for {
			select {
			case ev := <-f.watcher.Event:
				{
					if ev.IsDir() {
						continue
					}

					configName := cfile.GetFileName(ev.FileInfo.Name(), true)
					clog.Infof("Trigger file change. [name = %s]", configName)

					data, err := f.ReadBytes(configName)
					if err != nil {
						clog.Warn("Read data fail. [name = %s, err = %s]", configName, err)
						continue
					}

					if f.changeFn != nil {
						f.changeFn(configName, data)
					}
				}
			case err := <-f.watcher.Error:
				{
					clog.Error(err)
					continue
				}
			case <-f.watcher.Closed:
				return
			}
		}
	}()

	err := f.watcher.Start(time.Duration(f.ReloadTime) * time.Millisecond)
	if err != nil {
		clog.Panic(err)
	}
}

func (f *SourceFile) Stop() {
	if f.watcher == nil {
		return
	}

	err := f.watcher.Remove(f.monitorPath)
	clog.Infof("Remote watcher [path = %s, err = %v]", f.monitorPath, err)

	f.watcher.Close()
}

func (f *SourceFile) unmarshalFileConfig() error {
	//read data_config->file node
	dataConfig := cprofile.GetConfig("data_config").GetConfig(f.Name())

	err := dataConfig.Unmarshal(&f.fileConfig)
	if err != nil {
		return err
	}

	if f.ReloadTime < 1 {
		f.ReloadTime = 3000
	}

	if len(f.ExtName) < 1 {
		f.ExtName = ".json"
	}

	if len(f.FilePath) < 1 {
		f.FilePath = "data/"
	}

	return nil
}
