package cherryDataConfig

import (
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/radovskyb/watcher"
	"hash/crc32"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"
)

type FileSource struct {
	dataConfig IDataConfig

	monitorPath string //监控的路径
	filesCRC    map[string]uint32
	watcher     *watcher.Watcher
	reloadTime  int64
}

func (l *FileSource) Name() string {
	return "file"
}

func (l *FileSource) Init(dataConfig IDataConfig) {
	l.dataConfig = dataConfig

	if l.check() == false {
		return
	}

	for _, file := range dataConfig.GetFiles() {
		l.loadFile(file.FileName())
	}

	go l.newWatcher()
}

func (l *FileSource) check() bool {
	//read data_config->file node
	fileNode := cherryProfile.Config().Get("data_config", "file")
	if fileNode == nil {
		cherryLogger.Warnf("`data_config` node not found in `%s` file.", cherryProfile.FilePath())
		return false
	}

	filePath := fileNode.Get("file_path").ToString()
	if filePath == "" {
		filePath = "data_config/"
	}

	var err error
	l.monitorPath, err = cherryUtils.File.JoinPath(cherryProfile.Dir(), filePath)
	if err != nil {
		cherryLogger.Warn(err)
		return false
	}

	l.reloadTime = fileNode.Get("reload_time").ToInt64()
	if l.reloadTime < 1 {
		l.reloadTime = 2000
	}

	// init
	l.filesCRC = make(map[string]uint32)

	return true
}

func (l *FileSource) loadFile(fileName string) {
	if fileName == "" {
		cherryLogger.Warn("file name is empty.")
		return
	}

	fullPath, err := cherryUtils.File.JoinPath(l.monitorPath, fileName)
	if err != nil {
		cherryLogger.Warnf("file not found. err = %v path = %s", err, fullPath)
		return
	}

	bytes, err := ioutil.ReadFile(fullPath)
	if err != nil {
		cherryLogger.Warnf("read file err. err = %v path = %s", err, fullPath)
		return
	}

	if len(bytes) < 1 {
		return
	}

	newCrc := crc32.ChecksumIEEE(bytes)
	crcValue := l.filesCRC[fileName]

	if newCrc != crcValue {
		l.filesCRC[fileName] = newCrc
		l.dataConfig.Load(fileName, bytes)
		cherryLogger.Infof("[%s] file load complete.", fileName)
	}
}

func (l *FileSource) newWatcher() {
	l.watcher = watcher.New()
	l.watcher.SetMaxEvents(1)
	l.watcher.FilterOps(watcher.Write)

	err := l.watcher.AddRecursive(l.monitorPath)
	if err != nil {
		cherryLogger.Warn("new watcher error. path=%s, err=%v", l.monitorPath, err)
		return
	}

	//new goroutine
	go func() {
		for {
			select {
			case ev := <-l.watcher.Event:
				{
					fileName := filepath.Base(ev.Name())
					l.loadFile(fileName)
				}
			case err := <-l.watcher.Error:
				{
					cherryLogger.Error(err)
					return
				}
			case <-l.watcher.Closed:
				return
			}
		}
	}()

	if err := l.watcher.Start(time.Millisecond * time.Duration(l.reloadTime)); err != nil {
		cherryLogger.Warn(err)
	}
}

func (l *FileSource) Stop() {
	if l.watcher != nil {
		err := l.watcher.Remove(l.monitorPath)
		if err != nil {
			cherryLogger.Warn(err)
		}
		cherryLogger.Infof("remove watcher [path = %s]", l.monitorPath)
		l.watcher.Closed <- struct{}{}
	}
}

func (l *FileSource) getFullPath(fileName string) string {
	return path.Join(l.monitorPath, fileName)
}
