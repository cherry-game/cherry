package cherryDataConfig

import (
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/fsnotify/fsnotify"
	"hash/crc32"
	"io/ioutil"
	"path"
)

type LocalSource struct {
	parse           IDataParse
	filePath        string
	reloadFlushTime int
	filesCRC        map[string]uint32
	watch           *fsnotify.Watcher
	watchRunning    bool
}

func (l *LocalSource) Name() string {
	return "local"
}

func (l *LocalSource) Init() {
	fileNode := cherryProfile.Config().Get("dataConfig", "file")
	if fileNode == nil {
		cherryLogger.Error("`file` node info not found from settings.json")
		return
	}

	filePath := fileNode.Get("filePath").ToString()
	if filePath == "" {
		filePath = "/dataconfig"
	}
	l.filePath = path.Join(cherryProfile.ConfigPath(), filePath)

	reloadFlushTime := fileNode.Get("reloadFlushTime").ToInt()
	if reloadFlushTime < 1 {
		reloadFlushTime = 3000
	}
	l.reloadFlushTime = reloadFlushTime

	if l.reloadFlushTime > 0 {
		if l.watch == nil {
			l.watch, _ = fsnotify.NewWatcher()
		}

		l.watch.Add(l.filePath)

		l.watchRunning = true
		go l.watchEvent()
	}
}

func (l *LocalSource) watchEvent() {
	for l.watchRunning {
		select {
		case ev := <-l.watch.Events:
			{
				if ev.Op&fsnotify.Write == fsnotify.Write {
					//fmt.Println("写入文件 : ", ev.Name)
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

func (l *LocalSource) Destroy() {
	l.watchRunning = false
	l.watch.Remove(l.filePath)
}

func (l *LocalSource) SetParse(parse IDataParse) {
	l.parse = parse
}

func (l *LocalSource) GetContent(fileName string) ([]byte, error) {
	fullPath := l.getFullPath(fileName)
	bytes, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	l.filesCRC[fileName] = crc32.ChecksumIEEE(bytes)
	return bytes, nil
}

func (l *LocalSource) getFullPath(fileName string) string {
	return path.Join(l.filePath, fileName)
}

func (l *LocalSource) GetConfigNames() []string {
	keys := make([]string, 0, len(l.filesCRC))
	for k := range l.filesCRC {
		keys = append(keys, k)
	}
	return keys
}
