package event

type BaseEvent struct {
	name     string
	senderId string
}

func (*BaseEvent) Name() string {
	return ""
}

func (p *BaseEvent) SenderID() string {
	return p.senderId
}
