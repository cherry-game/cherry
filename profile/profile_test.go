package cherryProfile

import (
	"fmt"
	"testing"
)

func TestLoadFile(t *testing.T) {
	path := "../../examples/config/profile-dev.json"
	node, err := Init(path, "game-1")
	fmt.Println(node, err)
}
