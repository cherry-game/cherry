package cherryHandler

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"runtime/debug"
)

type (
	ExecutorEvent struct {
		groupIndex int
		Event      cfacade.IEvent
		EventSlice []cfacade.EventFunc
	}
)

func (p *ExecutorEvent) Index() int {
	return p.groupIndex
}

func (p *ExecutorEvent) SetIndex(index int) {
	p.groupIndex = index
}

func (p *ExecutorEvent) Invoke() {
	defer func() {
		if rev := recover(); rev != nil {
			clog.Warnf("recover in Event. %s", string(debug.Stack()))
			clog.Warnf("event = [%+v]", p.Event)
		}
	}()

	for _, fn := range p.EventSlice {
		fn(p.Event)
	}
}

func (p *ExecutorEvent) String() string {
	return p.Event.Name()
}
