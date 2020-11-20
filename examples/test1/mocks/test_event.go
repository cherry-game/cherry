package mocks

import (
	"github.com/phantacix/cherry/events"
)

type TestEvent struct {
	cherryEvents.GameEvent
	Abc int
}

func NewTestEvent() *TestEvent {
	return &TestEvent{
		GameEvent: cherryEvents.GameEvent{
			Name: "testEventName",
			Id:   "",
		},
		Abc: 0,
	}
}

func (t *TestEvent) EventName() string {
	return t.Name
}
