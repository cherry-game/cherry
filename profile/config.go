package cherryProfile

import (
	"time"

	cfacade "github.com/cherry-game/cherry/facade"
	jsoniter "github.com/json-iterator/go"
)

type (
	// Config wraps jsoniter.Any and provides type-safe config reading methods
	// with default value support.
	//
	// By embedding jsoniter.Any, Config inherits all native jsoniter methods
	// (Get, ToString, ToBool, etc.) while adding typed getters with fallback
	// defaults (GetString, GetBool, GetInt64, etc.) and an Unmarshal helper.
	//
	// Note: because jsoniter.Any is embedded, callers can still use its native
	// methods directly, which bypass Config's default-value logic. Prefer the
	// typed getter methods defined on Config.
	Config struct {
		jsoniter.Any
	}
)

// Wrap creates a Config from an arbitrary value by delegating to jsoniter.Wrap.
func Wrap(val any) *Config {
	return &Config{
		Any: jsoniter.Wrap(val),
	}
}

// GetConfig returns a sub-config at the given path as a ProfileJSON.
// Path semantics match jsoniter.Get.
func (p *Config) GetConfig(path ...any) cfacade.ProfileJSON {
	return &Config{
		Any: p.Any.Get(path...),
	}
}

// GetString reads a string value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns "" if no default is provided.
func (p *Config) GetString(path any, defaultVal ...string) string {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
	return result.ToString()
}

// GetBool reads a bool value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns false if no default is provided.
func (p *Config) GetBool(path any, defaultVal ...bool) bool {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}

		return false
	}

	return result.ToBool()
}

// GetInt reads an int value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns 0 if no default is provided.
func (p *Config) GetInt(path any, defaultVal ...int) int {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToInt()
}

// GetInt32 reads an int32 value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns 0 if no default is provided.
func (p *Config) GetInt32(path any, defaultVal ...int32) int32 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToInt32()
}

// GetInt64 reads an int64 value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns 0 if no default is provided.
func (p *Config) GetInt64(path any, defaultVal ...int64) int64 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToInt64()
}

// GetUint reads a uint value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns 0 if no default is provided.
func (p *Config) GetUint(path any, defaultVal ...uint) uint {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToUint()
}

// GetUint32 reads a uint32 value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns 0 if no default is provided.
func (p *Config) GetUint32(path any, defaultVal ...uint32) uint32 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToUint32()
}

// GetUint64 reads a uint64 value at path. Returns defaultVal[0] if the path is
// missing or invalid; returns 0 if no default is provided.
func (p *Config) GetUint64(path any, defaultVal ...uint64) uint64 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToUint64()
}

// GetFloat32 reads a float32 value at path. Returns defaultVal[0] if the path
// is missing or invalid; returns 0 if no default is provided.
func (p *Config) GetFloat32(path any, defaultVal ...float32) float32 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToFloat32()
}

// GetFloat64 reads a float64 value at path. Returns defaultVal[0] if the path
// is missing or invalid; returns 0 if no default is provided.
func (p *Config) GetFloat64(path any, defaultVal ...float64) float64 {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return result.ToFloat64()
}

// GetDuration reads an integer value at path and casts it to time.Duration.
// Returns defaultVal[0] if the path is missing or invalid; returns 0 if no
// default is provided.
//
// IMPORTANT: the returned value is a raw time.Duration (nanoseconds). Callers
// must multiply by the intended unit, e.g.:
//
//	delay := config.GetDuration("reconnect_delay", 1) * time.Second
func (p *Config) GetDuration(path any, defaultVal ...time.Duration) time.Duration {
	result := p.Get(path)
	if result.LastError() != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	return time.Duration(result.ToInt64())
}

// Unmarshal deserializes the current config value into the given target.
// Returns the underlying jsoniter error if the value is invalid.
func (p *Config) Unmarshal(value any) error {
	if p.LastError() != nil {
		return p.LastError()
	}
	return jsoniter.UnmarshalFromString(p.ToString(), value)
}
