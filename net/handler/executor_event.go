package cherryHandler

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"runtime/debug"
)

type (
	ExecutorEvent struct {
		Executor
		event      cfacade.IEvent
		eventSlice []cfacade.EventFn
	}
)

func (p *ExecutorEvent) Event() cfacade.IEvent {
	return p.event
}

func (p *ExecutorEvent) Invoke() {
	defer func() {
		if rev := recover(); rev != nil {
			clog.Warnf("recover in Event. %s", string(debug.Stack()))
			clog.Warnf("event = [%+v]", p.Event)
		}
	}()

	for _, fn := range p.eventSlice {
		fn(p.event)
	}
}

func (p *ExecutorEvent) QueueHash(queueNum int) int {
	return int(p.event.UniqueId() % int64(queueNum))
}
