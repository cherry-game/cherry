package cherryProfile

import (
	"fmt"
	"testing"
)

func TestLoadFile(t *testing.T) {
	path := "../../examples/config/"
	name := ""
	node, err := Init(path, name, "game-1")
	fmt.Println(node, err)
}
