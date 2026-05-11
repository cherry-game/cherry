package cherryActor

import (
	"strings"
	"time"

	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"go.uber.org/zap/zapcore"
)

/**
- Each Actor runs independently in a goroutine; all logic is serialized.
- Actor receives three message types: Local, Remote, and Event.
	- Each type has its own queue consumed in FIFO order.
	- Local: messages from game clients.
	- Remote: messages between Actors.
	- Event: pub/sub event messages.
- An Actor can create child Actors; child messages are routed by the parent.
- An Actor can create multiple timers for scheduled tasks.
- Cross-node Actor communication via cluster and discovery components.
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
		event            *actorEvent           // event handle
		timer            *actorTimer           // timer handle
		child            *actorChild           // child actor
		lastAt           int64                 // last process time (ms)
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
	case <-p.timer.C:
		{
			p.processTimer()
		}
	case <-p.close:
		{
			p.state = StopState
		}
	}

	return false
}

// processLocal pops and processes one message from the local mailbox.
//
// Message lifecycle: each Actor pops a message and recycles it via defer.
// When forwarding to a child, the same pointer is shared (ref counted).
// The last actor to Recycle() puts the message back into the pool.
func (p *Actor) processLocal() {
	m := p.localMail.Pop()
	if m == nil {
		return
	}
	defer m.Recycle()

	p.lastAt = time.Now().UnixMilli()

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
				clog.Warnf("child actor not found. target=%s", m.Target)
			}
		}
	} else {
		p.invokeFunc(p.localMail, p.App(), p.system.localInvokeFunc, m)
	}
}

// processRemote pops and processes one message from the remote mailbox.
//
// Message lifecycle: same rules as processLocal (see above).
func (p *Actor) processRemote() {
	m := p.remoteMail.Pop()
	if m == nil {
		return
	}
	defer m.Recycle()

	p.lastAt = time.Now().UnixMilli()

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
				clog.Warnf("child actor not found. target=%s", m.Target)
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

	p.lastAt = time.Now().UnixMilli()
	p.event.invokeFunc(eventData)
}

func (p *Actor) processTimer() {
	timerID := p.timer.Pop()
	if timerID < 1 {
		return
	}

	p.timer.invokeFunc(timerID)
}

func (p *Actor) invokeFunc(mb *mailbox, app cfacade.IApplication, fn cfacade.InvokeFunc, m *cfacade.Message) {
	funcInfo, found := mb.funcMap[m.FuncName]
	if !found {
		clog.Warnf("[%s] function not found. source=%s target=%s func=%s",
			mb.name,
			m.Source,
			m.Target,
			m.FuncName,
		)
		return
	}

	p.arrivalElapsed = p.lastAt - m.BuildTime
	if p.arrivalElapsed > p.system.arrivalTimeOut {
		clog.Warnf("[%s] message arrived in %dms (limit=%dms) source=%s target=%s func=%s",
			mb.name,
			p.arrivalElapsed,
			p.system.arrivalTimeOut,
			m.Source,
			m.Target,
			m.FuncName,
		)
	}

	defer func() {
		p.executionElapsed = time.Now().UnixMilli() - p.lastAt
		if p.executionElapsed > p.system.executionTimeout {
			clog.Warnf("[%s] message executed in %dms (limit=%dms) source=%s target=%s func=%s",
				mb.name,
				p.executionElapsed,
				p.system.executionTimeout,
				m.Source,
				m.Target,
				m.FuncName,
			)
		}

		if rev := recover(); rev != nil {
			clog.Errorf("[%s] invoke panic. source=%s target=%s func=%s type=%v",
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
	// If current actor is already a child, stop message processing.
	if p.path.IsChild() {
		clog.Warnf("[findChildActor] cannot create child from child. target=%s func=%s",
			m.Target,
			m.FuncName,
		)
		return nil, false
	}

	// Look up child actor.
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
	p.handler.OnInit()
	p.state = WorkerState
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

func (p *Actor) Call(targetPath, funcName string, arg any) int32 {
	return p.system.Call(p.path.String(), targetPath, funcName, arg)
}

func (p *Actor) CallWait(targetPath, funcName string, arg, reply any) int32 {
	return p.system.CallWait(p.path.String(), targetPath, funcName, arg, reply)
}

func (p *Actor) CallType(nodeType, actorID, funcName string, arg any) int32 {
	return p.system.CallType(nodeType, actorID, funcName, arg)
}

// LastAt second
func (p *Actor) LastAt() int64 {
	return p.lastAt
}

func (p *Actor) Exit() {
	p.close <- struct{}{}

	if clog.PrintLevel(zapcore.DebugLevel) {
		clog.Debugf("[Exit] path=%s", p.path)
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
	m.AddRef()
	p.remoteMail.Push(m)
}

func (p *Actor) PostLocal(m *cfacade.Message) {
	m.AddRef()
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
		lastAt:  time.Now().UnixMilli(),
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

	// spawn load!
	actorLoad, ok := handler.(IActorLoader)
	if ok {
		actorLoad.load(&thisActor)
	}

	c.wg.Add(1)

	return &thisActor, nil
}
