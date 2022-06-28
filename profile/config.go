package cherryProfile

import (
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

func (p *Config) GetString(path interface{}, val ...string) string {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(val) > 0 {
			return val[0]
		}
		return ""
	}
	return result.ToString()
}

func (p *Config) GetBool(path interface{}, val ...bool) bool {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(val) > 0 {
			return val[0]
		}

		return false
	}

	return result.ToBool()
}

func (p *Config) GetInt(path interface{}, val ...int) int {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(val) > 0 {
			return val[0]
		}
		return 0
	}

	return result.ToInt()
}

func (p *Config) GetInt32(path interface{}, val ...int32) int32 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(val) > 0 {
			return val[0]
		}
		return 0
	}

	return result.ToInt32()
}

func (p *Config) GetInt64(path interface{}, val ...int64) int64 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(val) > 0 {
			return val[0]
		}
		return 0
	}

	return result.ToInt64()
}
