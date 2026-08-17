package cherryActor

import (
	"time"

	ctimeWheel "github.com/cherry-game/cherry/extend/time_wheel"
	clog "github.com/cherry-game/cherry/logger"
)

type (
	actorTimer struct {
		queue          // queue
		thisActor      *Actor
		timerInvokeMap map[uint64]func() // key:timerID,value:business callback (invokeFunc)
	}
)

func newTimer(thisActor *Actor) actorTimer {
	return actorTimer{
		queue:          newQueue(),
		thisActor:      thisActor,
		timerInvokeMap: make(map[uint64]func()),
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

func (p *actorTimer) New(delay time.Duration, fn func()) ITimerHandle {
	return p.newTimerHandle(delay, fn, false)
}

func (p *actorTimer) NewOnce(delay time.Duration, fn func()) ITimerHandle {
	return p.newTimerHandle(delay, fn, true)
}

func (p *actorTimer) Add(delay time.Duration, fn func()) ITimerHandle {
	t := p.newTimerHandle(delay, fn, false)
	if t == nil {
		return nil
	}
	t.Start()
	return t
}

func (p *actorTimer) AddOnce(delay time.Duration, fn func()) ITimerHandle {
	t := p.newTimerHandle(delay, fn, true)
	if t == nil {
		return nil
	}
	t.Start()
	return t
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
	timer = p.thisActor.system.timeWheel.AddScheduleTimer(s, func() {
		p.Push(timer.ID())
	})

	if timer == nil {
		clog.Warnf("Build schedule fail. ITimerSchedule = %+v, fn = %+v", s, fn)
		return nil
	}

	p.addTimerInvoke(timer.ID(), fn)

	return timer
}

func (p *actorTimer) Remove(id uint64) {
	if _, found := p.timerInvokeMap[id]; found {
		p.thisActor.system.timeWheel.RemoveTimer(id)
		delete(p.timerInvokeMap, id)
	}
}

func (p *actorTimer) RemoveAll() {
	for id := range p.timerInvokeMap {
		p.thisActor.system.timeWheel.RemoveTimer(id)
		delete(p.timerInvokeMap, id)
	}
}

// newTimerHandle validates the parameters and builds a timer handle bound to
// the actor without starting it: the wheel callback only pushes the timer id
// into the actor queue, and the business callback is registered for the actor
// goroutine to invoke. Add/AddOnce call Start on it; New/NewOnce return it
// unstarted.
func (p *actorTimer) newTimerHandle(delay time.Duration, fn func(), once bool) ITimerHandle {
	if delay < ctimeWheel.DefaultTick || fn == nil {
		clog.Warnf("[Timer] parameter error. delay = %+v", delay)
		return nil
	}
	var t ITimerHandle
	t = p.thisActor.system.timeWheel.NewTimer(delay, func() {
		p.Push(t.ID())
	}, once)
	p.addTimerInvoke(t.ID(), fn)
	return t
}

func (p *actorTimer) addTimerInvoke(timerID uint64, fn func()) {
	p.timerInvokeMap[timerID] = fn
}

func (p *actorTimer) invokeFunc(timerID uint64) {
	fn, found := p.timerInvokeMap[timerID]
	if !found {
		return
	}

	defer func() {
		if rev := recover(); rev != nil {
			clog.Errorf("[%s] Timer invoke error. [timerID = %d, err = %+v]",
				p.thisActor.Path(),
				timerID,
				rev,
			)
		}
	}()

	fn()
}
