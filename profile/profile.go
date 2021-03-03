package cherryProfile

import (
	"fmt"
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/extend/utils"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"path"
)

var (
	configDir   string       // config root dir
	fileName    string       // profile fileName
	profileName string       // profile name
	profileJson jsoniter.Any // profile-x.json parse to json object
	debug       bool         // is debug
)

func Dir() string {
	return configDir
}

func Name() string {
	return profileName
}

func FileName() string {
	return fileName
}

func Debug() bool {
	return debug
}

func Config(path ...interface{}) jsoniter.Any {
	if len(path) > 0 {
		return profileJson.Get(path...)
	}
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

	configDir = judgePath
	profileName = profile
	fileName = fmt.Sprintf(cherryConst.ProfileNameFormat, profile)

	profileJson = loadProfileFile(configDir, fileName)
	if profileJson == nil {
		return nil, cherryUtils.Errorf("load profile file error. configPath = %s", configPath)
	}

	debug = true // default debug is true
	if profileJson.Get("debug") != nil {
		debug = profileJson.Get("debug").ToBool()
	}

	return profileJson, nil
}

func loadProfileFile(configPath string, profileFilePath string) jsoniter.Any {
	// read master json file
	masterBytes, err := ioutil.ReadFile(path.Join(configPath, profileFilePath))
	if err != nil {
		panic(err)
	}

	masterNode := jsoniter.Get(masterBytes)
	includeNode := masterNode.Get("include")
	if includeNode.LastError() != nil {
		// not found include node
		return masterNode
	}

	// merge include json file
	var maps = make(map[string]interface{})
	err = jsoniter.Unmarshal(masterBytes, &maps)
	if err != nil {
		panic(err)
	}

	// read include json file
	for i := 0; i < includeNode.Size(); i++ {
		nodeName := includeNode.Get(i).ToString()
		err := readMaps(path.Join(configPath, nodeName), maps)
		if err != nil {
			panic(err)
		}
	}

	json, err := jsoniter.Marshal(&maps)
	if err != nil {
		panic(err)
	}

	return jsoniter.Get(json)
}

func readMaps(includePath string, maps map[string]interface{}) error {
	bytes, err := ioutil.ReadFile(includePath)
	if err != nil {
		return err
	}

	return jsoniter.Unmarshal(bytes, &maps)
}
