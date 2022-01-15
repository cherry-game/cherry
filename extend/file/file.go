package cherryFile

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	str := GetCurrentPath()
	fmt.Sprintf(str)
}

func JudgePath(filePath string) (string, bool) {
	dir := GetMainFuncDir()
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

var (
	mainFuncDir = ""
)

func GetMainFuncDir() []string {
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

		dir = append(dir, lastLine[:lastIndex])
	}

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
