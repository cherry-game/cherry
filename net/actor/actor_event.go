package cherryActor

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

type actorEvent struct {
	thisActor *Actor                // parent
	queue                           // queue
	funcMap   map[string]IEventFunc // register event func map
}

func newEvent(thisActor *Actor) actorEvent {
	return actorEvent{
		thisActor: thisActor,
		queue:     newQueue(),
		funcMap:   make(map[string]IEventFunc),
	}
}

// Register 注册事件
// name 事件名
// fn 接收事件处理的函数
func (p *actorEvent) Register(name string, fn IEventFunc) {
	p.funcMap[name] = fn
}

// Unregister 注销事件
// name 事件名
func (p *actorEvent) Unregister(name string) {
	delete(p.funcMap, name)
}

func (p *actorEvent) Push(data cfacade.IEventData) {
	if _, found := p.funcMap[data.Name()]; found {
		p.queue.Push(data)
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

func (p *actorEvent) funcInvoke(data cfacade.IEventData) {
	fn, found := p.funcMap[data.Name()]
	if !found {
		clog.Warnf("Event not found. data = %+v", data)
		return
	}

	defer func() {
		if rev := recover(); rev != nil {
			clog.Errorf("[%s] Event invoke error. [data = %+v]",
				data,
			)
		}
	}()

	fn(data)
}

func (p *actorEvent) onStop() {
	p.funcMap = nil
	p.queue.Destroy()
	p.thisActor = nil
}
