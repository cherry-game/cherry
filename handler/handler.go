package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/utils"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
)

type (
	Handler struct {
		cherryInterfaces.AppContext
		Worker
		name             string                                // unique Handler name
		eventFn          map[string][]cherryInterfaces.EventFn // event func
		localHandlers    map[string]*cherryInterfaces.InvokeFn // local invoke Handler functions
		remoteHandlers   map[string]*cherryInterfaces.InvokeFn // remote invoke Handler functions
		handlerComponent *HandlerComponent                     // handler component
	}
)

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) SetName(name string) {
	h.name = name
}

func (h *Handler) PreInit() {

	if h.eventFn == nil {
		h.eventFn = make(map[string][]cherryInterfaces.EventFn)
	}

	if h.localHandlers == nil {
		h.localHandlers = make(map[string]*cherryInterfaces.InvokeFn)
	}

	if h.remoteHandlers == nil {
		h.remoteHandlers = make(map[string]*cherryInterfaces.InvokeFn)
	}

	h.handlerComponent = h.App().Find(cherryConst.HandlerComponent).(*HandlerComponent)
	if h.handlerComponent == nil {
		cherryLogger.Warn("not find HandlerComponent.")
	}
}

func (h *Handler) Init() {
}

func (h *Handler) AfterInit() {
	if h.queueSize < 1 {
		h.queueSize = 32767
	}

	if h.workerSize < 1 {
		h.workerSize = 1
	}

	//init chan
	h.messageChan = make([]chan interface{}, h.workerSize)
	for i := 0; i < int(h.workerSize); i++ {
		h.messageChan[i] = make(chan interface{}, h.queueSize)
	}

	if h.workerExecutorFn == nil {
		h.workerExecutorFn = DefaultWorkerExecutor
	}

	for i := 0; i < int(h.workerSize); i++ {
		go h.workerExecutorFn(h, i, h.messageChan[i])
	}
}

func (h *Handler) GetEvents() map[string][]cherryInterfaces.EventFn {
	return h.eventFn
}

func (h *Handler) GetEvent(name string) ([]cherryInterfaces.EventFn, bool) {
	events, found := h.eventFn[name]
	return events, found
}

func (h *Handler) GetLocals() map[string]*cherryInterfaces.InvokeFn {
	return h.localHandlers
}

func (h *Handler) GetLocal(funcName string) (*cherryInterfaces.InvokeFn, bool) {
	invoke, found := h.localHandlers[funcName]
	return invoke, found
}

func (h *Handler) GetRemotes() map[string]*cherryInterfaces.InvokeFn {
	return h.remoteHandlers
}

func (h *Handler) GetRemote(funcName string) (*cherryInterfaces.InvokeFn, bool) {
	invoke, found := h.remoteHandlers[funcName]
	return invoke, found
}

func (h *Handler) PutMessage(message interface{}) {
	if h.workerSize == 1 {
		h.messageChan[0] <- message
	} else {
		index := h.workerHashFn(message)
		if index > h.workerSize {
			index = 0
		}
		h.messageChan[index] <- message
	}
}

func (h *Handler) Stop() {
}

func (h *Handler) HandlerComponent() *HandlerComponent {
	return h.handlerComponent
}

func (h *Handler) RegisterLocals(sliceFn ...interface{}) {
	for _, fn := range sliceFn {
		funcName := cherryUtils.Reflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.RegisterLocal(funcName, fn)
	}
}

func (h *Handler) RegisterLocal(name string, fn interface{}) {
	f, err := cherryUtils.Reflect.GetInvokeFunc(name, fn)
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
		funcName := cherryUtils.Reflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.RegisterRemote(funcName, fn)
	}
}

func (h *Handler) RegisterRemote(name string, fn interface{}) {
	invokeFunc, err := cherryUtils.Reflect.GetInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	h.remoteHandlers[name] = invokeFunc

	cherryLogger.Debugf("[Handler = %s] register remote func name = %s, numIn = %d, numOut = %d",
		h.name, name, len(invokeFunc.InArgs), len(invokeFunc.OutArgs))
}

func (h *Handler) PostEvent(e cherryInterfaces.IEvent) {
	h.handlerComponent.PostEvent(e)
}

//RegisterEvent
func (h *Handler) RegisterEvent(eventName string, fn cherryInterfaces.EventFn) {
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
