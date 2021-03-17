package cherryDataConfig

import (
	jsoniter "github.com/json-iterator/go"
)

type ParserJson struct {
}

func (j *ParserJson) TypeName() string {
	return "json"
}

func (j *ParserJson) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}
