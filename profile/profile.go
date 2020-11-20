package cherryProfile

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/phantacix/cherry/const"
	"github.com/phantacix/cherry/utils"
	"io/ioutil"
	"os"
	"path"
)

var (
	configDir string       //config dir
	name      string       //profile name
	config    jsoniter.Any //profile-x.json parse to json object
	debug     bool         //is debug
)

func ConfigDir() string {
	return configDir
}

func Name() string {
	return name
}

func Debug() bool {
	return debug
}

func Config() jsoniter.Any {
	return config
}

func Init(configPath, profileName string) error {
	if configPath == "" {
		return cherryUtils.Error("configPath parameter is null.")
	}

	if profileName == "" {
		return cherryUtils.Error("profileName parameter is null.")
	}

	judgeDir, ok := judgeConfigPath(configPath)
	if !ok {
		return cherryUtils.ErrorFormat("configPath = %s not found.", configPath)
	}

	profileFilePath := path.Join(judgeDir, fmt.Sprintf(cherryConst.ProfileFileName, profileName))
	_, err := os.Stat(profileFilePath)
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadFile(profileFilePath)
	if err != nil {
		return err
	}

	name = profileName
	configDir = judgeDir
	config = jsoniter.Get(bytes)
	debug = config.Get("debug").ToBool()

	return nil
}

func judgeConfigPath(configPath string) (string, bool) {
	ok := cherryUtils.File.IsDir(configPath)
	if ok {
		return configPath, true
	}

	tmpPath := path.Join(cherryUtils.File.GetWorkPath(), configPath)
	ok = cherryUtils.File.IsDir(tmpPath)
	if ok {
		return tmpPath, true
	}

	tmpPath = path.Join(cherryUtils.File.GetMainFuncDir(), configPath)
	ok = cherryUtils.File.IsDir(tmpPath)
	if ok {
		return tmpPath, true
	}
	return "", false
}
