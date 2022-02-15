package cherryJson

import (
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
)

func ToJson(i interface{}) string {
	if i == nil {
		return ""
	}

	bytes, err := jsoniter.Marshal(i)
	if err != nil {
		return ""
	}

	return string(bytes)
}

func ReadMaps(includePath string, maps map[string]interface{}) error {
	bytes, err := ioutil.ReadFile(includePath)
	if err != nil {
		return err
	}

	err = jsoniter.Unmarshal(bytes, &maps)
	if err != nil {
		return err
	}
	return nil
}
