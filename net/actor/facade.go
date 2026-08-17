package cherryActor

import (
	"time"

	creflect "github.com/cherry-game/cherry/extend/reflect"
	cfacade "github.com/cherry-game/cherry/facade"
)

type (
	IActorLoader interface {
		load(actor *Actor)
	}
)

type (
	IEvent interface {
		Register(name string, fn IEventFunc, uniqueID ...int64)     // register event
		Registers(names []string, fn IEventFunc, uniqueID ...int64) // register multiple events
		Unregister(name string)                                     // unregister event
	}

	IEventFunc func(cfacade.IEventData) // event handler
)

type (
	IMailBox interface {
		Register(funcName string, fn interface{}) // register handler function
		GetFuncInfo(funcName string) (*creflect.FuncInfo, bool)
	}
)

type (
	ITimer interface {
		New(d time.Duration, fn func()) ITimerHandle                   // create a recurring timer without starting it; call Start() to start it
		NewOnce(d time.Duration, fn func()) ITimerHandle               // create a one-shot timer without starting it; call Start() to start it
		Add(d time.Duration, fn func()) ITimerHandle                   // add a recurring timer and start it
		AddOnce(d time.Duration, fn func()) ITimerHandle               // add a one-shot timer and start it
		AddFixedHour(hour, minute, second int, fn func()) ITimerHandle // add daily timer at fixed hour:minute:second
		AddFixedMinute(minute, second int, fn func()) ITimerHandle     // add hourly timer at fixed minute:second
		AddSchedule(s ITimerSchedule, f func()) ITimerHandle           // add timer with custom schedule
		Remove(id uint64)                                              // remove timer
		RemoveAll()                                                    // remove all timers
	}

	ITimerHandle interface {
		ID() uint64                      // unique timer id
		Start()                          // start or restart the timer
		Stop()                           // stop the timer (immediate: no new callback starts after it returns)
		IsOnce() bool                    // one-shot vs recurring
		IsRunning() bool                 // whether the timer is currently active
		SetNext(fn func() time.Duration) // set/replace the next-delay callback (recurring: per fire, one-shot: single fire); returns <=0 stops
	}

	ITimerSchedule interface {
		Next(time.Time) time.Time
	}
)
