// Package cherryConst holds framework-wide constants: version, logo, and separators.
package cherryConst

import (
	"testing"
)

// TestVersionPrint prints the framework logo for visual verification.
func TestVersionPrint(t *testing.T) {
	t.Log(GetLOGO())
}
