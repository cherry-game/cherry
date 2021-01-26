package cherryUtils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	str "strings"
)

type file struct {
}

func (f *file) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return true
	}
	return false
}

func (f *file) GetMainFuncDir() string {
	var buf [2 << 16]byte
	stack := string(buf[:runtime.Stack(buf[:], true)])

	lines := str.Split(str.TrimSpace(stack), "\n")
	lastLine := str.TrimSpace(lines[len(lines)-1])

	path := lastLine[:str.LastIndex(lastLine, "/")]
	return path
}

func (f *file) GetWorkPath() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}

func (f *file) CheckPath(path string) {
	_, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func (f *file) GetAllFile(dir string, s []string) []string {
	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := dir + "/" + fi.Name()
			s = f.GetAllFile(fullDir, s)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s
			}
		} else {
			fullName := path.Join(dir, fi.Name())
			s = append(s, fullName)
		}
	}
	return s
}
