package cherryActor

import (
	"strings"

	ctime "github.com/cherry-game/cherry/extend/time"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"go.uber.org/zap/zapcore"
)

/**
- 每个Actor独立运行在一个goroutine中，所有的逻辑都是串行处理
- Actor接收三种消息：本地消息(Local)、远程消息(Remote)、事件消息(Event)
	- 三种消息都有自己的队列(Queue)，每个队列依据FIFO原则进行消费
	- 本地消息(Local)，用于接收游戏客户端发送过来的本地消息
	- 远程消息(Remote)，用于Actor之间调用的远程消息
	- 事件消息(Event)，通过订阅/发布进行的事件消息
- Actor可以创建多个子Actor(ChildActor)，子Actor的消息由父Actor进行路由转发
- Actor可以创建多个定时器(Timer)进行定时业务的处理
- 通过cluster集群组件、discovery发现服务组件，进行跨节点的actor通信
*/

var (
	_nilActor = &Actor{}
)

var (
	InitState   State = 0
	WorkerState State = 1
	FreeState   State = 2
	StopState   State = 3
)

type (
	State int

	Actor struct {
		system           *System               // actor system
		path             *cfacade.ActorPath    // actor path
		state            State                 // actor state
		close            chan struct{}         // close flag
		handler          cfacade.IActorHandler // actor handler
		localMail        *mailbox              // local message mailbox
		remoteMail       *mailbox              // remote message mailbox
		event            *actorEvent           // event
		child            *actorChild           // child actor
		timer            *actorTimer           // timer
		lastAt           int64                 // last process time (count of seconds)
		arrivalElapsed   int64                 // arrival elapsed for message
		executionElapsed int64                 // execution elapsed for message
	}
)

func (p *Actor) run() {
	p.onInit()
	defer p.onStop()

	for {
		if p.loop() {
			break
		}
	}
}

func (p *Actor) loop() bool {
	if p.state == StopState {
		if p.localMail.Count() < 1 &&
			p.remoteMail.Count() < 1 &&
			p.event.Count() < 1 {
			return true
		}
	}

	select {
	case <-p.localMail.C:
		{
			p.processLocal()
		}
	case <-p.remoteMail.C:
		{
			p.processRemote()
		}
	case <-p.event.C:
		{
			p.processEvent()
		}
	case <-p.close:
		{
			p.state = StopState
		}
	}

	return false
}

func (p *Actor) processLocal() {
	m := p.localMail.Pop()
	if m == nil {
		return
	}

	p.lastAt = ctime.Now().ToSecond()

	next, invoke := p.handler.OnLocalReceived(m)
	if invoke {
		p.invokeFunc(p.localMail, p.App(), p.system.localInvokeFunc, m)
	}

	if !next {
		return
	}

	if m.TargetPath().IsChild() {
		if p.path.IsChild() {
			p.invokeFunc(p.localMail, p.App(), p.system.localInvokeFunc, m)
		} else {
			if childActor, foundChild := p.findChildActor(m); foundChild {
				childActor.PostLocal(m)
			} else {
				clog.Warnf("Child actor not found. path = %s", m.Target)
			}
		}
	} else {
		p.invokeFunc(p.localMail, p.App(), p.system.localInvokeFunc, m)
	}
}

func (p *Actor) processRemote() {
	m := p.remoteMail.Pop()
	if m == nil {
		return
	}

	p.lastAt = ctime.Now().ToSecond()

	next, invoke := p.handler.OnRemoteReceived(m)
	if invoke {
		p.invokeFunc(p.remoteMail, p.App(), p.system.remoteInvokeFunc, m)
	}

	if !next {
		return
	}

	if m.TargetPath().IsChild() {
		if p.path.IsChild() {
			p.invokeFunc(p.remoteMail, p.App(), p.system.remoteInvokeFunc, m)
		} else {
			if childActor, foundChild := p.findChildActor(m); foundChild {
				childActor.PostRemote(m)
			} else {
				clog.Warnf("Child actor not found. path = %s", m.Target)
			}
		}
	} else {
		p.invokeFunc(p.remoteMail, p.App(), p.system.remoteInvokeFunc, m)
	}
}

func (p *Actor) processEvent() {
	eventData := p.event.Pop()
	if eventData == nil {
		return
	}

	p.lastAt = ctime.Now().ToSecond()
	p.event.invokeFunc(eventData)
}

func (p *Actor) invokeFunc(mb *mailbox, app cfacade.IApplication, fn cfacade.InvokeFunc, m *cfacade.Message) {
	funcInfo, found := mb.funcMap[m.FuncName]
	if !found {
		clog.Warnf("[%s] Function not found. [source = %s, target = %s -> %s]",
			mb.name,
			m.Source,
			m.Target,
			m.FuncName,
		)
		return
	}

	p.arrivalElapsed = m.PostTime - m.BuildTime
	if p.arrivalElapsed > p.system.arrivalTimeOut {
		clog.Warnf("[%s] Invoke timeout.[path = %s -> %s -> %s, postTime = %d, buildTime = %d, arrival = %dms]",
			mb.name,
			m.Source,
			m.Target,
			m.FuncName,
			m.PostTime,
			m.BuildTime,
			p.arrivalElapsed,
		)
	}

	now := ctime.Now().ToMillisecond()

	defer func() {
		p.executionElapsed = ctime.Now().ToMillisecond() - now
		if p.executionElapsed > p.system.executionTimeout {
			clog.Warnf("[%s] Invoke timeout.[source = %s, target = %s->%s, execution = %dms]",
				mb.name,
				m.Source,
				m.Target,
				m.FuncName,
				p.executionElapsed,
			)
		}

		if rev := recover(); rev != nil {
			clog.Errorf("[%s] Invoke error. [source = %s, target = %s->%s, type = %v]",
				mb.name,
				m.Source,
				m.Target,
				m.FuncName,
				funcInfo.InArgs,
			)
		}
	}()

	fn(app, funcInfo, m)
}

func (p *Actor) findChildActor(m *cfacade.Message) (*Actor, bool) {
	// 如果当前actor为子actor,则终止本次消息处理
	if p.path.IsChild() {
		clog.Warnf("[findChildActor] Child actor cannot be created again。[target = %s->%s]",
			m.Target,
			m.FuncName,
		)
		return nil, false
	}

	// 寻找childActor
	childActor, found := p.child.Get(m.TargetPath().ChildID)
	if !found {
		childActor, found = p.handler.OnFindChild(m)
	}

	if found {
		if cActor, ok := childActor.(*Actor); ok {
			return cActor, true
		}
	}

	return nil, false
}

func (p *Actor) onInit() {
	p.state = WorkerState
	p.handler.OnInit()
}

func (p *Actor) onStop() {
	cutils.Try(func() {
		close(p.close)

		if p.path.IsParent() {
			p.system.removeActor(p.ActorID())
			p.child.onStop()
		} else {
			if parent, found := p.system.GetActor(p.path.ActorID); found {
				parent.child.Remove(p.path.ChildID)
			}
		}

		p.handler.OnStop()
		p.timer.onStop()
		p.event.onStop()
		p.localMail.onStop()
		p.remoteMail.onStop()
	}, func(errString string) {
		clog.Error(errString)
	})

	p.system.wg.Done()
}

func (p *Actor) State() State {
	return p.state
}

func (p *Actor) App() cfacade.IApplication {
	return p.system.app
}

func (p *Actor) ActorID() string {
	if p.path.IsChild() {
		return p.path.ChildID
	}

	return p.path.ActorID
}

func (p *Actor) Path() *cfacade.ActorPath {
	return p.path
}

func (p *Actor) PathString() string {
	return p.path.String()
}

func (p *Actor) Call(targetPath, funcName string, arg interface{}) int32 {
	return p.system.Call(p.path.String(), targetPath, funcName, arg)
}

func (p *Actor) CallWait(targetPath, funcName string, arg interface{}, reply interface{}) int32 {
	return p.system.CallWait(p.path.String(), targetPath, funcName, arg, reply)
}

// LastAt second
func (p *Actor) LastAt() int64 {
	return p.lastAt
}

func (p *Actor) Exit() {
	p.close <- struct{}{}

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[Exit] actor exit! path = %s", p.path)
	}
}

func (p *Actor) System() *System {
	return p.system
}

func (p *Actor) Local() IMailBox {
	return p.localMail
}

func (p *Actor) Remote() IMailBox {
	return p.remoteMail
}

func (p *Actor) Event() IEvent {
	return p.event
}

func (p *Actor) Child() cfacade.IActorChild {
	return p.child
}

func (p *Actor) Timer() ITimer {
	return p.timer
}

func (p *Actor) PostRemote(m *cfacade.Message) {
	p.remoteMail.Push(m)
}

func (p *Actor) PostLocal(m *cfacade.Message) {
	p.localMail.Push(m)
}

func (p *Actor) PostEvent(data cfacade.IEventData) {
	p.system.PostEvent(data)
}

func newActor(actorID, childID string, handler cfacade.IActorHandler, c *System) (*Actor, error) {
	if strings.TrimSpace(actorID) == "" {
		clog.Error("[newActor] actor id is nil.")
		return _nilActor, ErrActorIDIsNil
	}

	thisActor := Actor{
		path: &cfacade.ActorPath{
			NodeID:  c.NodeID(),
			ActorID: actorID,
			ChildID: childID,
		},
		state:   InitState,
		system:  c,
		close:   make(chan struct{}, 1),
		handler: handler,
		lastAt:  ctime.Now().ToSecond(),
	}

	localMailbox := newMailbox(LocalName)
	thisActor.localMail = &localMailbox

	remoteMailbox := newMailbox(RemoteName)
	thisActor.remoteMail = &remoteMailbox

	event := newEvent(&thisActor)
	thisActor.event = &event

	child := newChild(&thisActor)
	thisActor.child = &child

	timer := newTimer(&thisActor)
	thisActor.timer = &timer

	// register update timer func
	thisActor.remoteMail.Register(updateTimerFuncName, thisActor.timer._updateTimer_)

	// spawn load!
	actorLoad, ok := handler.(IActorLoader)
	if ok {
		actorLoad.load(&thisActor)
	}

	c.wg.Add(1)

	return &thisActor, nil
}
