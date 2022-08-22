package main

type TestEvent struct {
	Abc int32
}

func (p *TestEvent) Name() string {
	return "testEventName"
}

func (p *TestEvent) UniqueId() int64 {
	return int64(p.Abc)
}

func NewTestEvent() *TestEvent {
	return &TestEvent{
		Abc: 0,
	}
}
