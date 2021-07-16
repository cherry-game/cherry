package cherryResult

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestResultJSON(t *testing.T) {

	user := struct {
		Id int `json:"id"`
	}{
		Id: 1,
	}

	result := New()

	d, _ := json.Marshal(result)
	s := string(d)
	fmt.Println(s)

	result.Data = &user
	d, _ = json.Marshal(result)
	s = string(d)
	fmt.Println(s)

}
