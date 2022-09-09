package event

type ActorLogin struct {
	BaseActorEvent
	LoginDays       int //累计登录天数(每天+1)
	ContinueDays    int //连续登录天数(超过一天则归零)
	MaxContinueDays int //最大连续登录天数
}

func (*ActorLogin) Name() string {
	return ActorLoginKey
}

func NewActorLogin(actorId int64) *ActorLogin {
	event := &ActorLogin{
		BaseActorEvent: BaseActorEvent{
			Id: actorId,
		},
	}
	return event
}
