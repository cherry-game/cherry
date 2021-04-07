package mocks

import (
	"github.com/cherry-game/cherry/net/event"
)

type TestEvent struct {
	cherryEvent.GameEvent
	Abc int
}

func NewTestEvent() *TestEvent {
	return &TestEvent{
		GameEvent: cherryEvent.GameEvent{
			Name: "testEventName",
			Id:   "",
		},
		Abc: 0,
	}
}

func (t *TestEvent) EventName() string {
	return t.Name
}
