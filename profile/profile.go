package cherryProfile

import (
	"path/filepath"

	cerror "github.com/cherry-game/cherry/error"
	cfile "github.com/cherry-game/cherry/extend/file"
	cjson "github.com/cherry-game/cherry/extend/json"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
)

var (
	cfg = &struct {
		profilePath string  // profile root dir
		profileName string  // profile name
		jsonConfig  *Config // profile-x.json parse to json object
		env         string  // env name
		debug       bool    // debug default is true
		printLevel  string  // cherry log print level
	}{}
)

func Path() string {
	return cfg.profilePath
}

func Name() string {
	return cfg.profileName
}

func Env() string {
	return cfg.env
}

func Debug() bool {
	return cfg.debug
}

func PrintLevel() string {
	return cfg.printLevel
}

func Init(filePath, nodeID string) (cfacade.INode, error) {
	if filePath == "" {
		return nil, cerror.Error("File path is nil.")
	}

	if nodeID == "" {
		return nil, cerror.Error("NodeID is nil.")
	}

	judgePath, ok := cfile.JudgeFile(filePath)
	if !ok {
		return nil, cerror.Errorf("File path error. filePath = %s", filePath)
	}

	p, f := filepath.Split(judgePath)
	jsonConfig, err := LoadFile(p, f)
	if err != nil || jsonConfig.Any == nil || jsonConfig.LastError() != nil {
		return nil, cerror.Errorf("Load profile file error. [err = %v]", err)
	}

	node, err := GetNodeWithConfig(jsonConfig, nodeID)
	if err != nil {
		return nil, cerror.Errorf("Failed to get node config from profile file. [err = %v]", err)
	}

	// init cfg
	cfg.profilePath = p
	cfg.profileName = f
	cfg.jsonConfig = jsonConfig
	cfg.env = jsonConfig.GetString("env", "default")
	cfg.debug = jsonConfig.GetBool("debug", true)
	cfg.printLevel = jsonConfig.GetString("print_level", "debug")

	return node, nil
}

func GetConfig(path ...interface{}) cfacade.ProfileJSON {
	return cfg.jsonConfig.GetConfig(path...)
}

func LoadFile(filePath, fileName string) (*Config, error) {
	var (
		profileMaps = make(map[string]interface{})
		includeMaps = make(map[string]interface{})
		rootMaps    = make(map[string]interface{})
	)

	// read profile json file
	fileNamePath := filepath.Join(filePath, fileName)
	if err := cjson.ReadMaps(fileNamePath, profileMaps); err != nil {
		return nil, err
	}

	// read include json file
	if v, found := profileMaps["include"].([]interface{}); found {
		paths := cstring.ToStringSlice(v)
		for _, p := range paths {
			includePath := filepath.Join(filePath, p)
			if err := cjson.ReadMaps(includePath, includeMaps); err != nil {
				return nil, err
			}
		}
	}

	mergeMap(rootMaps, includeMaps)
	mergeMap(rootMaps, profileMaps)

	return Wrap(rootMaps), nil
}

func mergeMap(dst, src map[string]interface{}) {
	for key, value := range src {
		if v, ok := dst[key]; ok {
			if m1, ok := v.(map[string]interface{}); ok {
				if m2, ok := value.(map[string]interface{}); ok {
					mergeMap(m1, m2)
				} else {
					dst[key] = value
				}
			} else {
				dst[key] = value
			}
		} else {
			dst[key] = value
		}
	}
}
