package cherryConst

import (
	"github.com/cherry-game/cherry/logger"
	"testing"
)

func TestVersionPrint(t *testing.T) {
	cherryLogger.DefaultSet()

	PrintVersion()
}
