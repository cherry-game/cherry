package cherryActor

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

type actorEvent struct {
	thisActor *Actor                  // parent
	queue                             // queue
	funcMap   map[string][]IEventFunc // register event func map
}

func newEvent(thisActor *Actor) actorEvent {
	return actorEvent{
		thisActor: thisActor,
		queue:     newQueue(),
		funcMap:   make(map[string][]IEventFunc),
	}
}

// Register 注册事件
// name 事件名
// fn 接收事件处理的函数
func (p *actorEvent) Register(name string, fn IEventFunc) {
	funcList := p.funcMap[name]
	funcList = append(funcList, fn)
	p.funcMap[name] = funcList
}

func (p *actorEvent) Registers(names []string, fn IEventFunc) {
	for _, name := range names {
		p.Register(name, fn)
	}
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
	p.funcMap = nil
	p.queue.Destroy()
	p.thisActor = nil
}
