package cherryActor

import (
	"time"

	cherryTimeWheel "github.com/cherry-game/cherry/extend/time_wheel"
	cutils "github.com/cherry-game/cherry/extend/utils"
	clog "github.com/cherry-game/cherry/logger"
)

type (
	actorTimer struct {
		queue                              // queue
		thisActor    *Actor                // this actor
		timerInfoMap map[uint64]*timerInfo // key:timerID,value:*timerInfo
	}

	timerInfo struct {
		timer *cherryTimeWheel.Timer
		fn    func()
		once  bool
	}
)

func newTimer(thisActor *Actor) actorTimer {
	return actorTimer{
		queue:        newQueue(),
		thisActor:    thisActor,
		timerInfoMap: make(map[uint64]*timerInfo),
	}
}

func (p *actorTimer) onStop() {
	p.RemoveAll()
	p.thisActor = nil
}

func (p *actorTimer) Push(data uint64) {
	p.queue.Push(data)
}

func (p *actorTimer) Pop() uint64 {
	v := p.queue.Pop()
	if v == nil {
		return 0
	}

	timerID, ok := v.(uint64)
	if !ok {
		clog.Warnf("Convert to Timer ID fail. v = %+v", v)
		return 0
	}

	return timerID
}

func (p *actorTimer) Add(delay time.Duration, fn func(), async ...bool) uint64 {
	if delay.Milliseconds() < 1 || fn == nil {
		clog.Warnf("[ActorTimer] Add parameter error. delay = %+v", delay)
		return 0
	}

	newID := globalTimer.NextID()
	timer := globalTimer.AddEveryFunc(newID, delay, p.timerTrigger(newID), async...)

	if timer == nil {
		clog.Warnf("[Add] error. delay = %+v", delay)
		return 0
	}

	p.addTimerInfo(timer, fn, false)

	return newID
}

func (p *actorTimer) AddOnce(delay time.Duration, fn func(), async ...bool) uint64 {
	if delay.Milliseconds() < 1 || fn == nil {
		clog.Warnf("[AddOnce] parameter error. delay = %+v", delay)
		return 0
	}

	timerID := globalTimer.NextID()
	timer := globalTimer.AfterFunc(timerID, delay, p.timerTrigger(timerID), async...)

	if timer == nil {
		clog.Warnf("[AddOnce] error. d = %+v", delay)
		return 0
	}

	p.addTimerInfo(timer, fn, true)

	return timerID
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

	timerID := globalTimer.NextID()
	timer := globalTimer.ScheduleFunc(timerID, s, func() {
		p.Push(timerID)
	}, async...)

	p.addTimerInfo(timer, fn, false)

	return timerID
}

func (p *actorTimer) Remove(id uint64) {
	if funcItem, found := p.timerInfoMap[id]; found {
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

func (p *actorTimer) invokeFunc(timerID uint64) {
	value, found := p.timerInfoMap[timerID]
	if !found {
		return
	}

	cutils.Try(func() {
		value.fn()
	}, func(errString string) {
		clog.Error(errString)
	})

	if value.once {
		delete(p.timerInfoMap, timerID)
	}
}

func (p *actorTimer) timerTrigger(timerID uint64) func() {
	return func() {
		if p != nil {
			p.Push(timerID)
		}
	}
}
