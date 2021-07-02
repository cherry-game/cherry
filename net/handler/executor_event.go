package cherryHandler

import facade "github.com/cherry-game/cherry/facade"

type (
	EventExecutor struct {
		Event   facade.IEvent
		EventFn []facade.EventFn
	}
)

func (e *EventExecutor) Invoke() {
	for _, fn := range e.EventFn {
		fn(e.Event)
	}
}

func (e *EventExecutor) String() string {
	return e.Event.Name()
}
