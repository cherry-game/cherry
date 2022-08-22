package cherryProfile

import "testing"

func TestLoadFile(t *testing.T) {
	path := "../../examples/config/"
	name := "dev"
	Init(path, name)
}
