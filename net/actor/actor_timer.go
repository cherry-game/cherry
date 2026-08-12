package cherryActor

import (
	"time"

	ctimeWheel "github.com/cherry-game/cherry/extend/time_wheel"
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
		timer ITimerHandle
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

func (p *actorTimer) Add(delay time.Duration, fn func()) ITimerHandle {
	if delay < ctimeWheel.DefaultTick || fn == nil {
		clog.Warnf("[ActorTimer] Add parameter error. delay = %+v", delay)
		return nil
	}

	var timer ITimerHandle
	timer = p.thisActor.system.timeWheel.Add(delay, func() {
		p.Push(timer.ID())
	})

	if timer == nil {
		clog.Warnf("[Add] error. delay = %+v", delay)
		return nil
	}

	p.addTimerInfo(timer, fn, false)

	return timer
}

func (p *actorTimer) AddOnce(delay time.Duration, fn func()) ITimerHandle {
	if delay < ctimeWheel.DefaultTick || fn == nil {
		clog.Warnf("[AddOnce] parameter error. delay = %+v", delay)
		return nil
	}

	var timer ITimerHandle
	timer = p.thisActor.system.timeWheel.AddOnce(delay, func() {
		p.Push(timer.ID())
	})

	if timer == nil {
		clog.Warnf("[AddOnce] error. d = %+v", delay)
		return nil
	}

	p.addTimerInfo(timer, fn, true)

	return timer
}

func (p *actorTimer) AddFixedHour(hour, minute, second int, fn func()) ITimerHandle {
	schedule := &ctimeWheel.FixedDateSchedule{
		Hour:   hour,
		Minute: minute,
		Second: second,
	}

	return p.AddSchedule(schedule, fn)
}

func (p *actorTimer) AddFixedMinute(minute, second int, fn func()) ITimerHandle {
	return p.AddFixedHour(-1, minute, second, fn)
}

func (p *actorTimer) AddSchedule(s ITimerSchedule, fn func()) ITimerHandle {
	if s == nil || fn == nil {
		return nil
	}

	var timer ITimerHandle
	timer = p.thisActor.system.timeWheel.AddSchedule(s, func() {
		p.Push(timer.ID())
	})

	if timer == nil {
		return nil
	}

	p.addTimerInfo(timer, fn, false)

	return timer
}

func (p *actorTimer) RemoveAll() {
	for _, info := range p.timerInfoMap {
		info.timer.Stop()
	}
}

func (p *actorTimer) addTimerInfo(timer ITimerHandle, fn func(), once bool) {
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
