package cherryCron

import (
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"testing"
	"time"
)

func TestAddEveryDayFunc(t *testing.T) {
	AddEveryDayFunc(func() {
		cherryLogger.Info(cherryTime.Now().ToDateTimeFormat())
	}, 17, 32, 5)

	AddEveryHourFunc(func() {
		cherryLogger.Info(cherryTime.Now().ToDateTimeFormat())
		panic("print panic~~~")
	}, 5, 5)

	AddDurationFunc(func() {
		cherryLogger.Info(cherryTime.Now().ToDateTimeFormat())
	}, 1*time.Minute)

	Run()
}
