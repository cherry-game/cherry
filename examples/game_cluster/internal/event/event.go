package event

type BaseActorEvent struct {
	Id int64
}

func (*BaseActorEvent) Name() string {
	return ""
}

func (p *BaseActorEvent) UniqueId() int64 {
	return p.Id
}

func (p *BaseActorEvent) ActorId() int64 {
	return p.Id
}
