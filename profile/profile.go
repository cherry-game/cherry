package cherryProfile

import (
	"fmt"
	cerr "github.com/cherry-game/cherry/error"
	cfile "github.com/cherry-game/cherry/extend/file"
	cjson "github.com/cherry-game/cherry/extend/json"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
	"path/filepath"
)

var (
	env = &struct {
		profilePath string  // profile root dir
		profileName string  // profile name
		fileName    string  // profile fileName
		jsonConfig  *Config // profile-x.json parse to json object
		debug       bool    // debug default is true
		printLevel  string  // cherry log print level
	}{}
)

const (
	profilePrefix = "profile-"
	profileSuffix = ".json"
)

func Dir() string {
	return env.profilePath
}

func Name() string {
	return env.profileName
}

func FileName() string {
	return env.fileName
}

func Debug() bool {
	return env.debug
}

func PrintLevel() string {
	return env.printLevel
}

func Init(profilePath, profileName, nodeId string) (cfacade.INode, error) {
	if profilePath == "" {
		profilePath = "./config"
	}

	if nodeId == "" {
		return nil, cerr.Error("nodeId is nil")
	}

	judgePath, ok := cfile.JudgePath(profilePath)
	if !ok {
		return nil, cerr.Errorf("path error. profilePath = %s", profilePath)
	}

	fileNameList, err := judgeNameList(judgePath, profileName)
	if err != nil {
		return nil, err
	}

	for _, fileName := range fileNameList {
		cfg, err := loadFile(judgePath, fileName)
		if err != nil || cfg.Any == nil || cfg.LastError() != nil {
			continue
		}

		node, err := GetNodeWithConfig(cfg, nodeId)
		if err != nil {
			continue
		}

		// init env
		env.profilePath = judgePath
		env.profileName = profileName
		env.fileName = fileName
		env.jsonConfig = cfg
		env.debug = env.jsonConfig.GetBool("debug", true)
		env.printLevel = env.jsonConfig.GetString("print_level", "debug")

		return node, nil
	}

	return nil, cerr.Errorf("profile file not found. nodeId = %s", nodeId)
}

func GetConfig(path ...interface{}) cfacade.JsonConfig {
	return env.jsonConfig.GetConfig(path...)
}

func loadFile(profilePath string, profileFullName string) (*Config, error) {
	// merge include json file
	var maps = make(map[string]interface{})

	// read master json file
	fullPath := filepath.Join(profilePath, profileFullName)
	if err := cjson.ReadMaps(fullPath, maps); err != nil {
		return nil, err
	}

	// read include json file
	if v, found := maps["include"].([]interface{}); found {
		paths := cstring.ToStringSlice(v)
		for _, p := range paths {
			includePath := filepath.Join(profilePath, p)
			if err := cjson.ReadMaps(includePath, maps); err != nil {
				return nil, err
			}
		}
	}

	return Wrap(maps), nil
}

func judgeNameList(path, name string) ([]string, error) {
	var list []string

	if name != "" {
		fileName := mergeProfileName(name)
		list = append(list, fileName)

	} else {
		// find path
		filesPath, err := cfile.ReadDir(path, "profile-", ".json")
		if err != nil {
			return nil, err
		}

		if len(filesPath) < 1 {
			return nil, cerr.Errorf("[path = %s] profile file not found.", path)
		}

		for _, fp := range filesPath {
			list = append(list, fp)
		}
	}

	return list, nil
}

func mergeProfileName(name string) string {
	return fmt.Sprintf("%s%s%s", profilePrefix, name, profileSuffix)
}
