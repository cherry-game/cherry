package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"time"
)

type (
	Handler struct {
		facade.AppContext
		WorkerGroup
		name             string                       // unique name
		eventFn          map[string][]facade.EventFn  // event func
		localHandlers    map[string]*facade.HandlerFn // local invoke Handler functions
		remoteHandlers   map[string]*facade.HandlerFn // remote invoke Handler functions
		handlerComponent *Component                   // handler component
	}
)

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) SetName(name string) {
	h.name = name
}

func (h *Handler) OnPreInit() {
	if h.eventFn == nil {
		h.eventFn = make(map[string][]facade.EventFn)
	}

	if h.localHandlers == nil {
		h.localHandlers = make(map[string]*facade.HandlerFn)
	}

	if h.remoteHandlers == nil {
		h.remoteHandlers = make(map[string]*facade.HandlerFn)
	}

	h.handlerComponent = h.App().Find(cherryConst.HandlerComponent).(*Component)
	if h.handlerComponent == nil {
		cherryLogger.Warn("not found handlerComponent.")
	}
}

func (h *Handler) OnInit() {
}

func (h *Handler) OnAfterInit() {
	h.initWorkerGroup()
	h.runWorker(h)
}

func (h *Handler) Events() map[string][]facade.EventFn {
	return h.eventFn
}

func (h *Handler) Event(name string) ([]facade.EventFn, bool) {
	events, found := h.eventFn[name]
	return events, found
}

func (h *Handler) LocalHandlers() map[string]*facade.HandlerFn {
	return h.localHandlers
}

func (h *Handler) LocalHandler(funcName string) (*facade.HandlerFn, bool) {
	invoke, found := h.localHandlers[funcName]
	return invoke, found
}

func (h *Handler) RemoteHandlers() map[string]*facade.HandlerFn {
	return h.remoteHandlers
}

func (h *Handler) RemoteHandler(funcName string) (*facade.HandlerFn, bool) {
	invoke, found := h.remoteHandlers[funcName]
	return invoke, found
}

func (h *Handler) PostMessage(message interface{}) {
	if message == nil {
		cherryLogger.Warn("put message is nil")
		return
	}

	worker := h.GetWorker(message)
	if worker == nil {
		cherryLogger.Warnf("message %v  not found worker.", message)
		return
	}

	worker.MessageChan <- message
}

func (h *Handler) GetWorker(message interface{}) *Worker {
	if h.workerSize == 1 {
		return h.workerMap[0]
	}

	index := h.workerHashFn(message)
	if index > h.workerSize {
		index = 0
	}
	return h.workerMap[index]
}

func (h *Handler) OnStop() {
	for _, worker := range h.workerMap {
		for {
			size := len(worker.MessageChan)
			cherryLogger.Infof("[%s-chan-%d] waiting goroutine is empty. len(chan)=%d", h.name, worker.Index, size)

			if size == 0 {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (h *Handler) HandlerComponent() *Component {
	return h.handlerComponent
}

func (h *Handler) RegisterLocals(sliceFn ...interface{}) {
	for _, fn := range sliceFn {
		funcName := cherryReflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.RegisterLocal(funcName, fn)
	}
}

func (h *Handler) RegisterLocal(name string, fn interface{}) {
	f, err := cherryReflect.GetInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	h.localHandlers[name] = f

	cherryLogger.Debugf("[Handler = %s] register local func name = %s, numIn = %d, numOut =%d",
		h.name, name, len(f.InArgs), len(f.OutArgs))
}

func (h *Handler) RegisterRemotes(sliceFn ...interface{}) {
	for _, fn := range sliceFn {
		funcName := cherryReflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.RegisterRemote(funcName, fn)
	}
}

func (h *Handler) RegisterRemote(name string, fn interface{}) {
	invokeFunc, err := cherryReflect.GetInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	h.remoteHandlers[name] = invokeFunc

	cherryLogger.Debugf("[Handler = %s] register remote func name = %s, numIn = %d, numOut = %d",
		h.name, name, len(invokeFunc.InArgs), len(invokeFunc.OutArgs))
}

func (h *Handler) PostEvent(e facade.IEvent) {
	h.handlerComponent.PostEvent(e)
}

//RegisterEvent
func (h *Handler) RegisterEvent(eventName string, fn facade.EventFn) {
	if eventName == "" {
		cherryLogger.Warn("eventName is nil")
		return
	}

	if fn == nil {
		cherryLogger.Warn("event function is nil")
		return
	}

	events := h.eventFn[eventName]
	events = append(events, fn)

	h.eventFn[eventName] = events

	if cherryProfile.Debug() {
		cherryLogger.Debugf("[Handler = %s] register event = %s.", h.name, eventName)
	}
}
