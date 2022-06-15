package cherryHandler

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"runtime/debug"
)

type (
	ExecutorEvent struct {
		Event      facade.IEvent
		EventSlice []facade.EventFunc
	}
)

func (p *ExecutorEvent) Invoke() {
	defer func() {
		if rev := recover(); rev != nil {
			cherryLogger.Warnf("recover in Event. %s", string(debug.Stack()))
			cherryLogger.Warnf("event = [%+v]", p.Event)
		}
	}()

	for _, fn := range p.EventSlice {
		fn(p.Event)
	}
}

func (p *ExecutorEvent) String() string {
	return p.Event.Name()
}
