package main

import (
	"github.com/cherry-game/cherry"
	cherryTimer "github.com/cherry-game/cherry/component/timer"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/logger"
	"time"
)

func main() {
	testApp := cherry.NewApp("../config/", "local", "web-1")
	defer testApp.OnShutdown()

	timerComponent := cherryTimer.NewComponent()

	cherryTimer.NewTimer(func() {
		cherryLogger.Infof("execute func.... %s", cherryTime.Now().ToDateTimeFormat())
	}, 1*time.Second, 0)

	testApp.Startup(timerComponent)
}
