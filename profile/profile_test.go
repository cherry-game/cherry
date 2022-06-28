package cherryProfile

import "testing"

func TestLoadFile(t *testing.T) {
	path := "../_examples/config/"
	name := "dev"
	Init(path, name)
}
