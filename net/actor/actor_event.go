package cherryActor

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

type (
	actorEvent struct {
		thisActor *Actor                  //
		queue                             // queue
		funcMap   map[string][]IEventFunc // register event func list
	}
)

func newEvent(thisActor *Actor) actorEvent {
	return actorEvent{
		thisActor: thisActor,
		queue:     newQueue(),
		funcMap:   make(map[string][]IEventFunc), // make(map[string]*eventFuncList)
	}
}

// Register 注册事件
// name     事件名
// fn       接收事件处理的函数
// uniqueID match IEventData.UniqueID()
func (p *actorEvent) Register(name string, fn IEventFunc, uniqueID ...int64) {
	// add event to system actor
	p.thisActor.system.addActorEvent(p.thisActor.ActorID(), name, uniqueID...)

	// name bind func
	p.funcMap[name] = append(p.funcMap[name], fn)
}

func (p *actorEvent) Registers(names []string, fn IEventFunc, uniqueID ...int64) {
	for _, name := range names {
		p.Register(name, fn, uniqueID...)
	}
}

// Unregister 注销事件
// name 事件名
func (p *actorEvent) Unregister(name string) {
	p.thisActor.system.removeActorEvent(p.thisActor.ActorID(), name)
	delete(p.funcMap, name)
}

func (p *actorEvent) Push(data cfacade.IEventData) {
	p.queue.Push(data)
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

	for _, eventFunc := range funcList {
		eventFunc(data)
	}
}

func (p *actorEvent) onStop() {
	// remove event names
	p.thisActor.system.removeActorEvent(p.thisActor.ActorID(), p.EventNames()...)

	p.funcMap = nil
	p.queue.Destroy()
	p.thisActor = nil
}

func (p *actorEvent) EventNames() []string {
	var names []string
	for eventName := range p.funcMap {
		names = append(names, eventName)
	}
	return names
}
