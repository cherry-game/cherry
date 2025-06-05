package cherryActor

import (
	"time"

	cherryTimeWheel "github.com/cherry-game/cherry/extend/time_wheel"
	cutils "github.com/cherry-game/cherry/extend/utils"
	clog "github.com/cherry-game/cherry/logger"
)

const (
	updateTimerFuncName = "_updateTimer_"
)

type (
	actorTimer struct {
		thisActor    *Actor
		timerInfoMap map[uint64]*timerInfo //key:timerID,value:*timerInfo
	}

	timerInfo struct {
		timer *cherryTimeWheel.Timer
		fn    func()
		once  bool
	}
)

func newTimer(thisActor *Actor) actorTimer {
	return actorTimer{
		thisActor:    thisActor,
		timerInfoMap: make(map[uint64]*timerInfo),
	}
}

func (p *actorTimer) onStop() {
	p.RemoveAll()
	p.thisActor = nil
}

func (p *actorTimer) Add(delay time.Duration, fn func(), async ...bool) uint64 {
	if delay.Milliseconds() < 1 || fn == nil {
		clog.Warnf("[ActorTimer] Add parameter error. delay = %+v", delay)
		return 0
	}

	newID := globalTimer.NextID()
	timer := globalTimer.AddEveryFunc(newID, delay, p.callUpdateTimer(newID), async...)

	if timer == nil {
		clog.Warnf("[ActorTimer] Add error. delay = %+v", delay)
		return 0
	}

	p.addTimerInfo(timer, fn, false)

	return newID
}

func (p *actorTimer) AddOnce(delay time.Duration, fn func(), async ...bool) uint64 {
	if delay.Milliseconds() < 1 || fn == nil {
		clog.Warnf("[ActorTimer] AddOnce parameter error. delay = %+v", delay)
		return 0
	}

	newID := globalTimer.NextID()
	timer := globalTimer.AfterFunc(newID, delay, p.callUpdateTimer(newID), async...)

	if timer == nil {
		clog.Warnf("[ActorTimer] AddOnce error. d = %+v", delay)
		return 0
	}

	p.addTimerInfo(timer, fn, true)

	return newID
}

func (p *actorTimer) AddFixedHour(hour, minute, second int, fn func(), async ...bool) uint64 {
	schedule := &cherryTimeWheel.FixedDateSchedule{
		Hour:   hour,
		Minute: minute,
		Second: second,
	}

	return p.AddSchedule(schedule, fn, async...)
}

func (p *actorTimer) AddFixedMinute(minute, second int, fn func(), async ...bool) uint64 {
	return p.AddFixedHour(-1, minute, second, fn, async...)
}

func (p *actorTimer) AddSchedule(s ITimerSchedule, fn func(), async ...bool) uint64 {
	if s == nil || fn == nil {
		return 0
	}

	newID := globalTimer.NextID()
	timer := globalTimer.ScheduleFunc(newID, s, p.callUpdateTimer(newID), async...)

	p.addTimerInfo(timer, fn, false)

	return newID
}

func (p *actorTimer) Remove(id uint64) {
	funcItem, found := p.timerInfoMap[id]
	if found {
		funcItem.timer.Stop()
		delete(p.timerInfoMap, id)
	}
}

func (p *actorTimer) RemoveAll() {
	for _, info := range p.timerInfoMap {
		info.timer.Stop()
	}
}

func (p *actorTimer) addTimerInfo(timer *cherryTimeWheel.Timer, fn func(), once bool) {
	p.timerInfoMap[timer.ID()] = &timerInfo{
		timer: timer,
		fn:    fn,
		once:  once,
	}
}

func (p *actorTimer) callUpdateTimer(id uint64) func() {
	return func() {
		p.thisActor.Call(p.thisActor.PathString(), updateTimerFuncName, id)
	}
}

func (p *actorTimer) _updateTimer_(id uint64) {
	value, found := p.timerInfoMap[id]
	if !found {
		return
	}

	cutils.Try(func() {
		value.fn()
	}, func(errString string) {
		clog.Error(errString)
	})

	if value.once {
		delete(p.timerInfoMap, id)
	}
}
