package main

import (
	cherryString "github.com/cherry-game/cherry/extend/string"
)

type TestEvent struct {
	Abc int32
}

func (p *TestEvent) Name() string {
	return "testEventName"
}

func (p *TestEvent) UniqueId() string {
	return cherryString.Int32ToString(p.Abc)
}

func NewTestEvent() *TestEvent {
	return &TestEvent{
		Abc: 0,
	}
}
