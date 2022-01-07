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

	for i := 0; i <= 23; i++ {
		NewEveryDayTimer(func() {
			now := cherryTime.Now()
			cherryLogger.Infof("每天%d点%d分运行", now.Hour(), now.Minute())
		}, i, 12, 34)
		cherryLogger.Infof("添加 每天%d点执行的定时器", i)
	}

	for i := 0; i <= 59; i++ {
		NewEveryHourTimer(func() {
			cherryLogger.Infof("每小时%d分执行一次", cherryTime.Now().Minute())
		}, i, 0)
		cherryLogger.Infof("添加 每小时%d分的定时器", i)
	}

	wg.Wait()
}
