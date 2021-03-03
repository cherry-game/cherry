package cherryProfile

import (
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/extend/json"
	"github.com/cherry-game/cherry/extend/string"
	"github.com/cherry-game/cherry/extend/utils"
	jsoniter "github.com/json-iterator/go"
	"path"
)

var (
	env = &struct {
		dir      string       // config root dir
		profile  string       // profile name
		fileName string       // profile fileName
		json     jsoniter.Any // profile-x.json parse to json object
		debug    bool         // debug default is true
	}{}
)

func Dir() string {
	return env.dir
}

func Name() string {
	return env.profile
}

func FileName() string {
	return env.fileName
}

func Debug() bool {
	return env.debug
}

func Config(path ...interface{}) jsoniter.Any {
	if len(path) > 0 {
		return env.json.Get(path...)
	}
	return env.json
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

	env.debug = true
	env.dir = judgePath
	env.profile = profile
	env.fileName = fmt.Sprintf(cherryConst.ProfileNameFormat, profile)

	env.json = loadProfileFile(env.dir, env.fileName)
	if env.json == nil || env.json.LastError() != nil {
		return nil, cherryUtils.Errorf("load profile file error. configPath = %s", configPath)
	}

	if env.json.Get("debug").LastError() == nil {
		env.debug = env.json.Get("debug").ToBool()
	}

	return env.json, nil
}

func loadProfileFile(configPath string, profileFileName string) jsoniter.Any {
	// merge include json file
	var maps = make(map[string]interface{})

	// read master json file
	err := cherryJson.ReadMaps(path.Join(configPath, profileFileName), maps)
	if err != nil {
		panic(err)
	}

	// read include json file
	if v, found := maps["include"].([]interface{}); found {
		paths := cherryString.ToStringSlice(v)
		for _, p := range paths {
			includePath := path.Join(configPath, p)
			err := cherryJson.ReadMaps(includePath, maps)
			if err != nil {
				panic(err)
			}
		}
	}

	return jsoniter.Wrap(maps)
}
