package cherryConst

import (
	"github.com/phantacix/cherry/logger"
	"testing"
)

func TestVersionPrint(t *testing.T) {
	cherryLogger.DefaultSet()

	PrintVersion()
}
