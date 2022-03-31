package cherryProfile

import (
	"fmt"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFile "github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/extend/json"
	"github.com/cherry-game/cherry/extend/string"
	jsoniter "github.com/json-iterator/go"
	"path"
)

var (
	env = &struct {
		profilePath string       // profile root dir
		profileName string       // profile name
		fileName    string       // profile fileName
		json        jsoniter.Any // profile-x.json parse to json object
		debug       bool         // debug default is true
	}{}
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

func Init(profilePath, profileName string) (jsoniter.Any, error) {
	if profilePath == "" {
		return nil, cherryError.Error("profilePath parameter is null.")
	}

	if profileName == "" {
		return nil, cherryError.Error("profileName parameter is null.")
	}

	judgePath, ok := cherryFile.JudgePath(profilePath)
	if !ok {
		return nil, cherryError.Errorf("profilePath = %s not found.", profilePath)
	}

	env.debug = true
	env.profilePath = judgePath
	env.profileName = profileName
	env.fileName = fmt.Sprintf(cherryConst.ProfileNameFormat, profileName)

	env.json = loadProfileFile(env.profilePath, env.fileName)
	if env.json == nil || env.json.LastError() != nil {
		return nil, cherryError.Errorf("load profile file error. profilePath = %s", profilePath)
	}

	env.debug = GetBool(env.json, "debug", true)

	return env.json, nil
}

func Get(path ...interface{}) jsoniter.Any {
	return env.json.Get(path...)
}

func loadProfileFile(profilePath string, profileFullName string) jsoniter.Any {
	// merge include json file
	var maps = make(map[string]interface{})

	// read master json file
	err := cherryJson.ReadMaps(path.Join(profilePath, profileFullName), maps)
	if err != nil {
		panic(err)
	}

	// read include json file
	if v, found := maps["include"].([]interface{}); found {
		paths := cherryString.ToStringSlice(v)
		for _, p := range paths {
			includePath := path.Join(profilePath, p)
			err := cherryJson.ReadMaps(includePath, maps)
			if err != nil {
				panic(err)
			}
		}
	}

	return jsoniter.Wrap(maps)
}
