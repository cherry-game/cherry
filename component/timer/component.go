package cherryTimer

import (
	cherryConst "github.com/cherry-game/cherry/const"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
)

type Component struct {
	cherryFacade.Component
	close chan bool
}

func NewComponent() *Component {
	return &Component{
		close: make(chan bool),
	}
}

func (p *Component) Name() string {
	return cherryConst.TimerComponent
}

func (p *Component) Init() {
	go func() {
		for {
			select {
			case <-p.close:
				break
			default:
				Cron()
			}
		}
	}()
	cherryLogger.Info("timer component init.")
}

func (p *Component) OnBeforeStop() {
	p.close <- true
}

func (p *Component) OnStop() {
	timerCount := 0
	Manager.timers.Range(func(idInterface, tInterface interface{}) bool {
		t := tInterface.(*Timer)
		t.Stop()

		timerCount++
		return true
	})
	cherryLogger.Infof("clean all timer.... [count = %d]", timerCount)
}
