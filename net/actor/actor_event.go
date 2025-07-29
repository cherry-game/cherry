package cherryActor

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

type (
	actorEvent struct {
		thisActor *Actor                    //
		queue                               // queue
		funcMap   map[string]*eventFuncList // register event func list
	}

	eventFuncList struct {
		list     []IEventFunc
		isUnique bool
		uniqueID int64
	}
)

func newEvent(thisActor *Actor) actorEvent {
	return actorEvent{
		thisActor: thisActor,
		queue:     newQueue(),
		funcMap:   make(map[string]*eventFuncList),
	}
}

// Register 注册事件
// name     事件名
// fn       接收事件处理的函数
// uniqueID match IEventData.UniqueID()
func (p *actorEvent) Register(name string, fn IEventFunc, uniqueID ...int64) {
	funcList, found := p.funcMap[name]
	if !found {
		funcList = &eventFuncList{}
		p.funcMap[name] = funcList
	}

	funcList.list = append(funcList.list, fn)

	// If a unique ID is set, it will be matched when receiving events
	funcList.isUnique = len(uniqueID) > 0
	if funcList.isUnique {
		funcList.uniqueID = uniqueID[0]
	}
}

func (p *actorEvent) Registers(names []string, fn IEventFunc, uniqueID ...int64) {
	for _, name := range names {
		p.Register(name, fn, uniqueID...)
	}
}

// Unregister 注销事件
// name 事件名
func (p *actorEvent) Unregister(name string) {
	delete(p.funcMap, name)
}

func (p *actorEvent) Push(data cfacade.IEventData) {
	if funcList, found := p.funcMap[data.Name()]; found {
		if funcList.isUnique {
			if funcList.uniqueID == data.UniqueID() {
				p.queue.Push(data)
			}
		} else {
			p.queue.Push(data)
		}
	}

	if p.thisActor.Path().IsChild() {
		return
	}

	p.thisActor.Child().Each(func(iActor cfacade.IActor) {
		if childActor, ok := iActor.(*Actor); ok {
			childActor.event.Push(data)
		}
	})
}

func (p *actorEvent) Pop() cfacade.IEventData {
	v := p.queue.Pop()
	if v == nil {
		return nil
	}

	eventData, ok := v.(cfacade.IEventData)
	if !ok {
		clog.Warnf("Convert to IEventData fail. v = %+v", v)
		return nil
	}

	return eventData
}

func (p *actorEvent) invokeFunc(data cfacade.IEventData) {
	funcList, found := p.funcMap[data.Name()]
	if !found {
		clog.Warnf("[%s] Event not found. [data = %+v]",
			p.thisActor.Path(),
			data,
		)
		return
	}

	defer func() {
		if rev := recover(); rev != nil {
			clog.Errorf("[%s] Event invoke error. [data = %+v]",
				p.thisActor.Path(),
				data,
			)
		}
	}()

	for _, eventFunc := range funcList.list {
		eventFunc(data)
	}
}

func (p *actorEvent) onStop() {
	p.funcMap = nil
	p.queue.Destroy()
	p.thisActor = nil
}
