// Package cherryProfile provides profile configuration file loading and node config
// resolution for the Cherry framework.
//
// Core responsibilities:
//   - Load profile JSON config files (with include file merging)
//   - Resolve per-node configuration (node identity, address, settings)
//   - Provide type-safe config reading via the ProfileJSON interface
//
// Note: this package uses package-level global state (cfg). Only one Application
// instance per process is supported. Multiple instances will overwrite each other's
// global config.
package cherryProfile

import (
	"path/filepath"

	cerror "github.com/cherry-game/cherry/error"
	cfile "github.com/cherry-game/cherry/extend/file"
	cjson "github.com/cherry-game/cherry/extend/json"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
)

// cfg is the package-level global config singleton.
// Written once by Init(), then read-only by the getter functions.
// Not goroutine-safe: Init() must complete during startup before any reads.
var (
	cfg = &struct {
		profilePath string  // absolute path to the profile config directory
		profileName string  // profile config filename (e.g. "dev.json")
		jsonConfig  *Config // merged JSON config tree (main + includes)
		env         string  // environment name (e.g. "dev", "test", "prod")
		debug       bool    // debug mode flag, defaults to true
		printLevel  string  // cherry log output level, defaults to "debug"
	}{}
)

// Path returns the absolute path to the profile config directory.
func Path() string {
	return cfg.profilePath
}

// Name returns the profile config filename (e.g. "dev.json").
func Name() string {
	return cfg.profileName
}

// Env returns the current environment name (e.g. "dev", "test", "prod").
func Env() string {
	return cfg.env
}

// Debug returns whether debug mode is enabled.
func Debug() bool {
	return cfg.debug
}

// PrintLevel returns the cherry log output level (e.g. "debug", "info", "warn", "error").
func PrintLevel() string {
	return cfg.printLevel
}

// Init loads the profile config file and returns the configuration for the
// specified node.
//
// filePath is the path to the profile JSON file, nodeID is the target node
// identifier.
//
// Loading steps:
//  1. Read the main profile config file
//  2. Merge files referenced by the "include" field (include keys are overridden
//     by the main config)
//  3. Search the "node" section for a node matching nodeID
//
// The returned INode provides the node's address, type, settings, etc.
// Also initializes the package-level global config (cfg) for use by Path, Name,
// Env, Debug, PrintLevel, and GetConfig.
func Init(filePath, nodeID string) (cfacade.INode, error) {
	if filePath == "" {
		return nil, cerror.Error("file path is empty")
	}

	if nodeID == "" {
		return nil, cerror.Error("nodeID is empty")
	}

	judgePath, ok := cfile.JudgeFile(filePath)
	if !ok {
		return nil, cerror.Errorf("invalid file path: %s", filePath)
	}

	p, f := filepath.Split(judgePath)
	jsonConfig, err := LoadFile(p, f)
	if err != nil || jsonConfig.Any == nil || jsonConfig.LastError() != nil {
		return nil, cerror.Errorf("failed to load profile file: %v", err)
	}

	node, err := GetNodeWithConfig(jsonConfig, nodeID)
	if err != nil {
		return nil, cerror.Errorf("node config not found in profile file: %v", err)
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

// GetConfig reads a sub-config from the global config tree at the given path.
// Path semantics match jsoniter.Get. Must be called after Init(), otherwise
// cfg.jsonConfig is nil and this will panic.
func GetConfig(path ...any) cfacade.ProfileJSON {
	return cfg.jsonConfig.GetConfig(path...)
}

// LoadFile loads and merges profile config files.
//
// Merge strategy:
//  1. Read the main config file (fileName) into profileMaps
//  2. Read files listed in the "include" field of the main config into includeMaps
//  3. Merge includeMaps into rootMaps first, then merge profileMaps into rootMaps
//     Keys in the main config override matching keys from include files (deep merge)
//
// Returns the merged Config object.
func LoadFile(filePath, fileName string) (*Config, error) {
	var (
		profileMaps = make(map[string]any)
		includeMaps = make(map[string]any)
		rootMaps    = make(map[string]any)
	)

	// read profile json file
	fileNamePath := filepath.Join(filePath, fileName)
	if err := cjson.ReadMaps(fileNamePath, profileMaps); err != nil {
		return nil, err
	}

	// read include json file
	if v, found := profileMaps["include"].([]any); found {
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

// mergeMap deep-merges src into dst. When both dst and src have a map value for
// the same key, the merge recurses; otherwise src's value overwrites dst's.
func mergeMap(dst, src map[string]any) {
	for key, value := range src {
		if v, ok := dst[key]; ok {
			if m1, ok := v.(map[string]any); ok {
				if m2, ok := value.(map[string]any); ok {
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
