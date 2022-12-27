package cherrySerializer

import (
	jsoniter "github.com/json-iterator/go"
)

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

// Marshal returns the JSON encoding of v.
func (j *JSON) Marshal(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	return jsoniter.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func (j *JSON) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}

// Name returns the name of the serializer.
func (j *JSON) Name() string {
	return "json"
}
