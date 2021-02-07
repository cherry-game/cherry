package cherryFile

import (
	"os"
	"path"
	"runtime"
	"strings"
)

func JudgePath(filePath string) (string, bool) {
	ok := IsDir(filePath)
	if ok {
		return filePath, true
	}

	tmpPath := path.Join(GetWorkPath(), filePath)
	ok = IsDir(tmpPath)
	if ok {
		return tmpPath, true
	}

	tmpPath = path.Join(GetMainFuncDir(), filePath)
	ok = IsDir(tmpPath)
	if ok {
		return tmpPath, true
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

func GetMainFuncDir() string {
	var buf [2 << 16]byte
	stack := string(buf[:runtime.Stack(buf[:], true)])

	lines := strings.Split(strings.TrimSpace(stack), "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])

	filePath := lastLine[:strings.LastIndex(lastLine, "/")]
	return filePath
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
		return "", err
	}
	return filePath, nil
}

func CheckPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

//func GetAllFile(dir string, s []string) []string {
//	rd, err := ioutil.ReadDir(dir)
//	if err != nil {
//		fmt.Println("read dir fail:", err)
//		return s
//	}
//
//	for _, fi := range rd {
//		if fi.IsDir() {
//			fullDir := dir + "/" + fi.Name()
//			s = GetAllFile(fullDir, s)
//			if err != nil {
//				fmt.Println("read dir fail:", err)
//				return s
//			}
//		} else {
//			fullName := path.Join(dir, fi.Name())
//			s = append(s, fullName)
//		}
//	}
//	return s
//}
