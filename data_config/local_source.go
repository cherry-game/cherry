package cherryDataConfig

import (
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/fsnotify/fsnotify"
	"hash/crc32"
	"io/ioutil"
	"path"
	"path/filepath"
)

type FileSource struct {
	dataConfig IDataConfig

	monitorPath  string //监控的路径
	filesCRC     map[string]uint32
	watch        *fsnotify.Watcher
	watchRunning bool
}

func (l *FileSource) Name() string {
	return "file"
}

func (l *FileSource) Init(dataConfig IDataConfig) {
	l.filesCRC = make(map[string]uint32)
	l.dataConfig = dataConfig

	if l.check() == false {
		return
	}

	for _, file := range dataConfig.GetFiles() {
		l.loadFile(file.FileName())
	}

	l.newWatcher()
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
		cherryLogger.Warnf("read file err. err = %v, path = %s", err, fullPath)
		return
	}

	l.filesCRC[fileName] = crc32.ChecksumIEEE(bytes)
	l.dataConfig.Load(fileName, bytes)
	cherryLogger.Infof("[%s] file load complete.", fileName)
}

func (l *FileSource) check() bool {
	//read data_config->file node
	fileNode := cherryProfile.Config().Get("data_config", "file")
	if fileNode == nil {
		cherryLogger.Warnf("`data_config` node not found in `%s` file.",
			cherryProfile.FilePath())
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

	return true
}

func (l *FileSource) newWatcher() {
	if l.watch == nil {
		l.watch, _ = fsnotify.NewWatcher()
	}

	l.watch.Add(l.monitorPath)
	l.watchRunning = true

	//new goroutine
	go l.watchEvent()
}

func (l *FileSource) Destroy() {
	l.watchRunning = false
	l.watch.Remove(l.monitorPath)
}

func (l *FileSource) watchEvent() {
	for l.watchRunning {
		select {
		case ev := <-l.watch.Events:
			{
				if ev.Op&fsnotify.Write == fsnotify.Write {
					cherryLogger.Infof("%s file change", ev.Name)
					fileName := filepath.Base(ev.Name)
					l.loadFile(fileName)
				}
			}
		case err := <-l.watch.Errors:
			{
				cherryLogger.Error(err)
				return
			}
		}
	}
}

func (l *FileSource) getFullPath(fileName string) string {
	return path.Join(l.monitorPath, fileName)
}
