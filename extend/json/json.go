package cherryJson

import encodingJson "encoding/json"

func ToJson(i interface{}) string {
	bytes, err := encodingJson.Marshal(i)
	if err != nil {
		return ""
	}
	return string(bytes)
}
