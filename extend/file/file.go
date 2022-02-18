package cherryFile

import (
	cherrySlice "github.com/cherry-game/cherry/extend/slice"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func JudgeFile(filePath string) (string, bool) {
	if filePath == "" {
		return filePath, false
	}

	var p, n string
	index := strings.LastIndex(filePath, "/")
	if index > 0 {
		p = filePath[0:index]
		n = filePath[index+1:]
	} else {
		p = "./"
		n = filePath
	}

	dir := GetStackDir()
	for _, d := range dir {
		tmpPath := path.Join(d, p, n)
		if CheckPath(tmpPath) == nil {
			return tmpPath, true
		}
	}

	return path.Join(p, n), false
}

func JudgePath(filePath string) (string, bool) {
	dir := GetStackDir()
	for _, d := range dir {
		tmpPath := path.Join(d, filePath)
		ok := IsDir(tmpPath)
		if ok {
			return tmpPath, true
		}
	}

	tmpPath := path.Join(GetWorkPath(), filePath)
	ok := IsDir(tmpPath)
	if ok {
		return tmpPath, true
	}

	tmpPath = path.Join(GetCurrentDirectory(), filePath)
	ok = IsDir(tmpPath)
	if ok {
		return tmpPath, true
	}

	ok = IsDir(filePath)
	if ok {
		return filePath, true
	}

	return "", false
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return true
	}
	return false
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}

func GetCurrentPath() string {
	var absPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		absPath = path.Dir(filename)
	}
	return absPath
}

func GetStackDir() []string {
	var dir []string

	var buf [2 << 16]byte
	stack := string(buf[:runtime.Stack(buf[:], true)])
	lines := strings.Split(strings.TrimSpace(stack), "\n")

	for _, line := range lines {
		lastLine := strings.TrimSpace(line)
		lastIndex := strings.LastIndex(lastLine, "/")
		if lastIndex < 1 {
			continue
		}

		thisDir := lastLine[:lastIndex]
		if _, err := os.Stat(thisDir); err != nil {
			continue
		}

		if cherrySlice.StringIn(thisDir, dir) {
			continue
		}

		dir = append(dir, thisDir)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dir)))

	return dir
}

func GetWorkPath() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}

func JoinPath(elem ...string) (string, error) {
	filePath := path.Join(elem...)

	err := CheckPath(filePath)
	if err != nil {
		return filePath, err
	}
	return filePath, nil
}

func CheckPath(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	return err
}

func GetFileName(filePath string, removeExt bool) string {
	fileName := path.Base(filePath)
	if removeExt == false {
		return fileName
	}

	var suffix string
	suffix = path.Ext(fileName)

	return strings.TrimSuffix(fileName, suffix)
}
