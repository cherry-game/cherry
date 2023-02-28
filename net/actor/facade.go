package cherryActor

import (
	cfacade "github.com/cherry-game/cherry/facade"
	"time"
)

type (
	IActorLoader interface {
		load(actor Actor)
	}
)

type (
	IEvent interface {
		Register(name string, fn IEventFunc) // 注册事件
		Unregister(name string)              // 注销事件
	}

	IEventFunc func(cfacade.IEventData) // 接收事件数据时的处理函数
)

type (
	IMailBox interface {
		Register(funcName string, fn interface{}) // 注册执行函数
	}
)

type (
	ITimer interface {
		Add(cmd func(dt time.Duration), endAt time.Time, count ...int) (id string)      // 根据结束时间添加定时器
		AddEveryDay(cmd func(dt time.Duration), hour, minutes, seconds int) (id string) // 添加每天定时器
		AddEveryHour(cmd func(dt time.Duration), minutes, seconds int) (id string)      // 添加每小时定时器
		AddDuration(cmd func(dt time.Duration), duration time.Duration) (id string)     // 根据指定的时间截添加定时器
		Remove(id string)                                                               // 移除定时器
	}
)

type (
	ITask interface {
		RunTask(fn func())
	}
)
