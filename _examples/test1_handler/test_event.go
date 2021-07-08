package main

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
			EventName: "testEventName",
			Id:        "",
		},
		Abc: 0,
	}
}
