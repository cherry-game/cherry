package cherryProfile

import (
	"encoding/json"
	cherryError "github.com/cherry-game/cherry/error"
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cfacade "github.com/cherry-game/cherry/facade"
	jsoniter "github.com/json-iterator/go"
)

type (
	Config struct {
		jsoniter.Any
	}
)

func Wrap(val interface{}) *Config {
	cfg := &Config{
		Any: jsoniter.Wrap(val),
	}
	return cfg
}

func (p *Config) GetConfig(path ...interface{}) cfacade.JsonConfig {
	return &Config{
		Any: p.Any.Get(path...),
	}
}

func (p *Config) GetString(path interface{}, defaultVal ...string) string {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
	return result.ToString()
}

func (p *Config) GetBool(path interface{}, defaultVal ...bool) bool {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}

		return false
	}

	return result.ToBool()
}

func (p *Config) GetInt(path interface{}, defaultVal ...int) int {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToInt()
}

func (p *Config) GetInt32(path interface{}, defaultVal ...int32) int32 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToInt32()
}

func (p *Config) GetInt64(path interface{}, defaultVal ...int64) int64 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToInt64()
}

func (p *Config) GetJsonObject(path interface{}, ptrVal interface{}) error {
	str := p.GetString(path, "")
	if str == "" {
		return cherryError.Error("get path value is nil.")
	}

	if creflect.IsPtr(ptrVal) == false {
		return cherryError.Error("ptrVal type error.")
	}

	bytes := []byte(str)
	if json.Valid(bytes) == false {
		return cherryError.Error("value convert to bytes is error.")
	}

	err := json.Unmarshal(bytes, ptrVal)
	if err != nil {
		return err
	}

	return nil
}
