package cherryActor

import "time"

type actorTimer struct {
	thisActor   *Actor
	timerIdList []string
}

func newTimer(thisActor *Actor) actorTimer {
	return actorTimer{
		thisActor: thisActor,
	}
}

func (p *actorTimer) onStop() {
}

func (p *actorTimer) Add(cmd func(dt time.Duration), endAt time.Time, count ...int) (id string) {

	p.thisActor.Call(p.thisActor.PathString(), "_update_timer", "1")

	return ""
}

func (*actorTimer) AddEveryDay(cmd func(dt time.Duration), hour, minutes, seconds int) (id string) {
	return ""
}

func (*actorTimer) AddEveryHour(cmd func(dt time.Duration), minutes, seconds int) (id string) {
	return ""
}

func (*actorTimer) AddDuration(cmd func(dt time.Duration), duration time.Duration) (id string) {
	return ""
}

func (*actorTimer) Remove(id string) {

}
