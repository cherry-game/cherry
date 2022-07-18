package cherryProfile

import (
	"fmt"
	cconst "github.com/cherry-game/cherry/const"
	cerr "github.com/cherry-game/cherry/error"
	cfile "github.com/cherry-game/cherry/extend/file"
	cjson "github.com/cherry-game/cherry/extend/json"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
	"path"
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

func Init(profilePath, profileName string) error {
	if profilePath == "" {
		return cerr.Error("profilePath parameter is null.")
	}

	if profileName == "" {
		return cerr.Error("profileName parameter is null.")
	}

	judgePath, ok := cfile.JudgePath(profilePath)
	if !ok {
		return cerr.Errorf("profilePath = %s not found.", profilePath)
	}

	fileName := fmt.Sprintf(cconst.ProfileNameFormat, profileName)
	env.jsonConfig = loadFile(judgePath, fileName)
	if env.jsonConfig.Any == nil || env.jsonConfig.LastError() != nil {
		return cerr.Errorf("load profile file error. profilePath = %s", profilePath)
	}

	env.profilePath = judgePath
	env.profileName = profileName
	env.fileName = fileName
	env.debug = env.jsonConfig.GetBool("debug", true)
	env.printLevel = env.jsonConfig.GetString("print_level", "debug")
	return nil
}

func GetConfig(path ...interface{}) cfacade.JsonConfig {
	return env.jsonConfig.GetConfig(path...)
}

func loadFile(profilePath string, profileFullName string) *Config {
	// merge include json file
	var maps = make(map[string]interface{})

	// read master json file
	err := cjson.ReadMaps(path.Join(profilePath, profileFullName), maps)
	if err != nil {
		panic(err)
	}

	// read include json file
	if v, found := maps["include"].([]interface{}); found {
		paths := cstring.ToStringSlice(v)
		for _, p := range paths {
			includePath := path.Join(profilePath, p)
			if err := cjson.ReadMaps(includePath, maps); err != nil {
				panic(err)
			}
		}
	}

	return Wrap(maps)
}
