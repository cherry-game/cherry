package cherryProfile

import (
	"fmt"
	"regexp"
	"testing"
)

func TestLoadFile(t *testing.T) {

	regex, err := regexp.Compile("master-1")
	fmt.Println(regex, err)

	//path := "../../examples/config/demo-cluster.json"
	path := "./dev.json"
	gate1, err := Init(path, "master-1")
	fmt.Println(gate1, err)

	game1, err := Init(path, "1")
	fmt.Println(game1, err)
}
