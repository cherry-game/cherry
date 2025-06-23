package cherryProfile

import (
	"fmt"
	"testing"
)

func TestLoadFile(t *testing.T) {
	path := "../../examples/config/demo-cluster.json"
	gate1, err := Init(path, "gc-gate-1")
	fmt.Println(gate1, err)

	game1, err := Init(path, "1")
	fmt.Println(game1, err)
}
