package cherryTimer

import (
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"sync"
	"testing"
)

func TestNewEveryDayTimer(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	comp := NewComponent()
	comp.Init()

	NewEveryDayTimer(func() {
		now := cherryTime.Now()
		cherryLogger.Infof("每天%d点%d分运行", now.Hour(), now.Minute())
	}, 10, 11, 0)

	for i := 54; i <= 59; i++ {
		NewEveryHourTimer(func() {
			cherryLogger.Infof("每小时第%d分执行一次", cherryTime.Now().Minute())
		}, i, 0)
	}

	wg.Wait()
}
