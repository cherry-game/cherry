package event

type ActorCreate struct {
	BaseActorEvent
	ActorName string
	gender    int32
}

func (*ActorCreate) Name() string {
	return ActorCreateKey
}

func NewActorCreate(actorId int64, actorName string, gender int32) *ActorCreate {
	event := &ActorCreate{
		BaseActorEvent: BaseActorEvent{
			Id: actorId,
		},
		ActorName: actorName,
		gender:    gender,
	}
	return event
}
