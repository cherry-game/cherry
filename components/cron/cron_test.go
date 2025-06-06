package cherryCron

import (
	"testing"
	"time"

	ctime "github.com/cherry-game/cherry/extend/time"
	clog "github.com/cherry-game/cherry/logger"
)

func TestAddEveryDayFunc(t *testing.T) {
	AddEveryDayFunc(func() {
		now := ctime.Now()
		clog.Info(now.ToDateTimeFormat())
	}, 17, 32, 5)

	AddEveryHourFunc(func() {
		now := ctime.Now()
		clog.Info(now.ToDateTimeFormat())
		panic("print panic~~~")
	}, 5, 5)

	AddDurationFunc(func() {
		now := ctime.Now()
		clog.Info(now.ToDateTimeFormat())
	}, 1*time.Minute)

	Run()
}
