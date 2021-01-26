package cherrySerializer

import "encoding/json"

type JSONSerializer struct{}

func NewJSON() *JSONSerializer {
	return &JSONSerializer{}
}

// Marshal returns the JSON encoding of v.
func (j *JSONSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func (j *JSONSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Name returns the name of the serializer.
func (j *JSONSerializer) Name() string {
	return "json"
}
