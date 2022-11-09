package cherryDataConfig

import (
	cerr "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/extend/file"
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/radovskyb/watcher"
	"os"
	"time"
)

type (
	// SourceFile 本地读取数据配置文件
	SourceFile struct {
		fileConfig
		watcher  *watcher.Watcher
		changeFn ConfigChangeFn
	}

	fileConfig struct {
		FilePath    string `json:"file_path"`   // 配置文件路径
		ExtName     string `json:"ext_name"`    // 文件扩展名
		ReloadTime  int64  `json:"reload_time"` // 定时重载扫描(毫秒)
		MonitorPath string `json:"-"`           // 监控路径
	}
)

func (f *SourceFile) Name() string {
	return "file"
}

func (f *SourceFile) Init(_ IDataConfig) {
	//read data_config->file node
	dataConfig := cprofile.GetConfig("data_config").GetConfig(f.Name())
	if dataConfig.Marshal(&f.fileConfig) != nil {
		clog.Warnf("[data_config]->[%s] node in `%s` file not found.", f.Name(), cprofile.FileName())
		return
	}

	err := f.check()
	if err != nil {
		clog.Warn(err)
		return
	}

	// new watcher
	go f.newWatcher()
}

func (f *SourceFile) ReadBytes(configName string) (data []byte, error error) {
	if configName == "" {
		return nil, cerr.Error("configName is empty.")
	}

	fullPath, err := cherryFile.JoinPath(f.MonitorPath, configName+f.ExtName)
	if err != nil {
		return nil, cerr.Errorf("file not found. err = %v", err)
	}

	if cherryFile.IsDir(fullPath) {
		return nil, cerr.Errorf("path is dir. fullPath = %s", fullPath)
	}

	data, err = os.ReadFile(fullPath)
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

	if err := f.watcher.Add(f.MonitorPath); err != nil {
		clog.Warn("new watcher error. path=%s, err=%v", f.MonitorPath, err)
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
					clog.Infof("[name = %s] trigger file change.", configName)

					data, err := f.ReadBytes(configName)
					if err != nil {
						clog.Warn("[name = %s] read data error = %s", configName, err)
						return
					}

					if f.changeFn != nil {
						f.changeFn(configName, data)
					}
				}
			case err := <-f.watcher.Error:
				{
					clog.Error(err)
					return
				}
			case <-f.watcher.Closed:
				return
			}
		}
	}()

	if err := f.watcher.Start(time.Millisecond * time.Duration(f.ReloadTime)); err != nil {
		clog.Warn(err)
	}
}

func (f *SourceFile) Stop() {
	if f.watcher == nil {
		return
	}

	err := f.watcher.Remove(f.MonitorPath)
	clog.Warnf("remote watcher [path = %s, err = %v]", f.MonitorPath, err)
	//f.watcher.Closed <- struct{}{}
}

func (f *fileConfig) check() error {
	if len(f.ExtName) < 1 {
		f.ExtName = ".json"
	}

	if f.ReloadTime < 1 {
		f.ReloadTime = 3000
	}

	if len(f.FilePath) < 1 {
		f.FilePath = "data/"
	}

	var err error
	f.MonitorPath, err = cherryFile.JoinPath(cprofile.Dir(), f.FilePath)
	if err != nil {
		return err
	}

	return nil
}
