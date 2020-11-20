package cherryDataConfig

import jsoniter "github.com/json-iterator/go"

type Json struct {
}

func (j *Json) Parse(text []byte, v interface{}) error {
	return jsoniter.Unmarshal(text, v)
}
