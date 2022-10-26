package cherryHandler

import (
	"context"
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cmessage "github.com/cherry-game/cherry/net/message"
	csession "github.com/cherry-game/cherry/net/session"
	"reflect"
)

type (
	Handler struct {
		cfacade.AppContext
		name                 string                         // unique name
		eventFuncMap         map[string]*cfacade.EventInfo  // event func
		localHandlerFuncMap  map[string]*cfacade.MethodInfo // local invoke Handler functions
		remoteHandlerFuncMap map[string]*cfacade.MethodInfo // remote invoke Handler functions
		handlerComponent     *Component                     // handler component
	}
)

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) SetName(name string) {
	h.name = name
}

func (h *Handler) OnPreInit() {
	h.eventFuncMap = make(map[string]*cfacade.EventInfo)
	h.localHandlerFuncMap = make(map[string]*cfacade.MethodInfo)
	h.remoteHandlerFuncMap = make(map[string]*cfacade.MethodInfo)

	var found = false
	h.handlerComponent, found = h.Find(Name).(*Component)
	if found == false {
		panic("handler component not found.")
	}
}

func (h *Handler) OnInit() {
}

func (h *Handler) OnAfterInit() {
}

func (h *Handler) OnStop() {
}

func (h *Handler) Events() map[string]*cfacade.EventInfo {
	return h.eventFuncMap
}

func (h *Handler) Event(name string) (*cfacade.EventInfo, bool) {
	events, found := h.eventFuncMap[name]
	return events, found
}

func (h *Handler) LocalHandlers() map[string]*cfacade.MethodInfo {
	return h.localHandlerFuncMap
}

func (h *Handler) LocalHandler(funcName string) (*cfacade.MethodInfo, bool) {
	invoke, found := h.localHandlerFuncMap[funcName]
	return invoke, found
}

func (h *Handler) RemoteHandlers() map[string]*cfacade.MethodInfo {
	return h.remoteHandlerFuncMap
}

func (h *Handler) RemoteHandler(funcName string) (*cfacade.MethodInfo, bool) {
	invoke, found := h.remoteHandlerFuncMap[funcName]
	return invoke, found
}

func (h *Handler) Component() *Component {
	return h.handlerComponent
}

func (h *Handler) AddLocals(localFns ...interface{}) {
	for _, fn := range localFns {
		funcName := creflect.GetFuncName(fn)
		if funcName == "" {
			clog.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.AddLocal(funcName, fn)
	}
}

func (h *Handler) AddLocal(name string, fn interface{}, hashFn ...cfacade.QueueHashFn) {
	invokeFn, err := getInvokeFunc(name, fn)
	if err != nil {
		clog.Warn(err)
		return
	}

	if len(hashFn) > 0 {
		invokeFn.QueueHash = hashFn[0]
	}

	h.localHandlerFuncMap[name] = invokeFn
}

func (h *Handler) AddLocalWithNames(names []string, fn interface{}, hashFn ...cfacade.QueueHashFn) {
	for _, name := range names {
		h.AddLocal(name, fn, hashFn...)
	}
}

func (h *Handler) AddRemotes(remoteFns ...interface{}) {
	for _, fn := range remoteFns {
		funcName := creflect.GetFuncName(fn)
		if funcName == "" {
			clog.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.AddRemote(funcName, fn)
	}
}

func (h *Handler) AddRemote(name string, fn interface{}, hashFn ...cfacade.QueueHashFn) {
	invokeFn, err := getInvokeFunc(name, fn)
	if err != nil {
		clog.Warn(err)
		return
	}

	if len(hashFn) > 0 {
		invokeFn.QueueHash = hashFn[0]
	}

	h.remoteHandlerFuncMap[name] = invokeFn
}

func (h *Handler) AddRemoteWithNames(names []string, fn interface{}, hashFn ...cfacade.QueueHashFn) {
	for _, name := range names {
		h.AddRemote(name, fn, hashFn...)
	}
}

func getInvokeFunc(name string, fn interface{}) (*cfacade.MethodInfo, error) {
	invokeFunc, err := creflect.GetInvokeFunc(name, fn)
	if err != nil {
		return invokeFunc, err
	}

	if len(invokeFunc.InArgs) == 3 {
		if invokeFunc.InArgs[2] == reflect.TypeOf(&cmessage.Message{}) {
			invokeFunc.IsRaw = true
		}
	}

	return invokeFunc, err
}

func (h *Handler) AddEvent(eventName string, fn cfacade.EventFn, hashFn ...cfacade.QueueHashFn) {
	if eventName == "" {
		clog.Warn("eventName is nil")
		return
	}

	if fn == nil {
		clog.Warn("event function is nil")
		return
	}

	eventInfo, found := h.eventFuncMap[eventName]
	if found == false {
		eventInfo = &cfacade.EventInfo{}
	}

	eventInfo.List = append(eventInfo.List, fn)

	if len(hashFn) > 0 {
		eventInfo.QueueHash = hashFn[0]
	}

	h.eventFuncMap[eventName] = eventInfo
}

func (h *Handler) PostEvent(e cfacade.IEvent) {
	if h.handlerComponent == nil {
		clog.Errorf("handlerComponent is nil. [event = %s]", e)
		return
	}

	h.handlerComponent.PostEvent(e)
}

func (h *Handler) AddBeforeFilter(beforeFilters ...FilterFn) {
	if h.handlerComponent != nil {
		h.handlerComponent.AddBeforeFilter(beforeFilters...)
	}
}

func (h *Handler) AddAfterFilter(afterFilters ...FilterFn) {
	if h.handlerComponent != nil {
		h.handlerComponent.AddAfterFilter(afterFilters...)
	}
}

func (h *Handler) Response(ctx context.Context, session *csession.Session, data interface{}) {
	session.Response(ctx, data)
}
