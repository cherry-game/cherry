package cherryUtils

import encodingJson "encoding/json"

type json struct {
}

func (j *json) ToJson(i interface{}) string {
	bytes, err := encodingJson.Marshal(i)
	if err != nil {
		return ""
	}
	return string(bytes)
}
