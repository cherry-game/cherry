package cherryDataConfig

import jsoniter "github.com/json-iterator/go"

type JsonParser struct {
}

func (j *JsonParser) Name() string {
	return "json"
}

func (j *JsonParser) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}
