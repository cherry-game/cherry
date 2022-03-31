package cherryProfile

import jsoniter "github.com/json-iterator/go"

func GetString(json jsoniter.Any, path string, defaultValue ...string) string {
	result := json.Get(path)
	if result.LastError() != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}

	return result.ToString()
}

func GetBool(json jsoniter.Any, path string, defaultValue ...bool) bool {
	result := json.Get(path)
	if result.LastError() != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return false
	}

	return result.ToBool()
}

func GetInt(json jsoniter.Any, path string, defaultValue ...int) int {
	result := json.Get(path)
	if result.LastError() != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result.ToInt()
}

func GetInt32(json jsoniter.Any, path string, defaultValue ...int32) int32 {
	result := json.Get(path)
	if result.LastError() != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result.ToInt32()
}

func GetInt64(json jsoniter.Any, path string, defaultValue ...int64) int64 {
	result := json.Get(path)
	if result.LastError() != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result.ToInt64()
}
