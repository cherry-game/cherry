package cherryActor

import "time"

type actorTimer struct {
	timerMap map[string]func() // register timer map
}

func newTimer() actorTimer {
	return actorTimer{
		timerMap: make(map[string]func()),
	}
}

func (p *actorTimer) onStop() {
	p.timerMap = nil
}

func (*actorTimer) Add(cmd func(dt time.Duration), endAt time.Time, count ...int) (id string) {
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
