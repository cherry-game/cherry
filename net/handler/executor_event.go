package cherryHandler

import facade "github.com/cherry-game/cherry/facade"

type (
	ExecutorEvent struct {
		Event      facade.IEvent
		EventSlice []facade.EventFunc
	}
)

func (e *ExecutorEvent) Invoke() {
	for _, fn := range e.EventSlice {
		fn(e.Event)
	}
}

func (e *ExecutorEvent) String() string {
	return e.Event.Name()
}
