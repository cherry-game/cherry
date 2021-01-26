package main

import (
	"github.com/cherry-game/cherry/extend/timer"
	"github.com/cherry-game/cherry/logger"
	"time"
)

func main() {
	t := cherryTimer.NewTimer(func() {
		//time.Sleep(time.Second * 3)
		cherryLogger.Infof("execute func....")
	}, time.Second*1, 3)

	cherryTimer.AddTimer(t)

	go func() {
		for {
			cherryTimer.Cron()
		}
	}()

	time.Sleep(time.Second * 10000)
}
