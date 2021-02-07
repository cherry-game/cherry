package cherryProfile

import (
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/extend/utils"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"os"
	"path"
)

var (
	configDir       string       // config root dir
	profileFullPath string       // profile file full path
	profileName     string       // profile profileName
	profileJson     jsoniter.Any // profile-x.json parse to json object
	debug           bool         // is debug
)

func Dir() string {
	return configDir
}

func FilePath() string {
	return profileFullPath
}

func Name() string {
	return profileName
}

func Debug() bool {
	return debug
}

func Config() jsoniter.Any {
	return profileJson
}

func Init(configPath, profile string) (jsoniter.Any, error) {
	if configPath == "" {
		return nil, cherryUtils.Error("configPath parameter is null.")
	}

	if profile == "" {
		return nil, cherryUtils.Error("profile parameter is null.")
	}

	judgePath, ok := cherryFile.JudgePath(configPath)
	if !ok {
		return nil, cherryUtils.Errorf("configPath = %s not found.", configPath)
	}

	profileFilePath := path.Join(judgePath, fmt.Sprintf(cherryConst.ProfileNameFormat, profile))
	_, err := os.Stat(profileFilePath)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(profileFilePath)
	if err != nil {
		return nil, err
	}

	configDir = judgePath
	profileFullPath = profileFilePath
	profileName = profile

	profileJson = jsoniter.Get(bytes)
	debug = true

	if profileJson.Get("debug") != nil {
		debug = profileJson.Get("debug").ToBool()
	}

	return profileJson, nil
}
