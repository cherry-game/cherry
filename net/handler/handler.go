package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/session"
)

type (
	Handler struct {
		facade.AppContext
		name             string                       // unique name
		eventFn          map[string][]facade.EventFn  // event func
		localHandlers    map[string]*facade.HandlerFn // local invoke Handler functions
		remoteHandlers   map[string]*facade.HandlerFn // remote invoke Handler functions
		handlerComponent *Component                   // handler component
		sessionComponent *cherrySession.Component     // session component
	}
)

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) SetName(name string) {
	h.name = name
}

func (h *Handler) OnPreInit() {
	h.eventFn = make(map[string][]facade.EventFn)
	h.localHandlers = make(map[string]*facade.HandlerFn)
	h.remoteHandlers = make(map[string]*facade.HandlerFn)

	var found = false
	h.handlerComponent, found = h.App().Find(cherryConst.HandlerComponent).(*Component)
	if found == false {
		panic("handler handlerComponent not found.")
	}

	h.sessionComponent, found = h.App().Find(cherryConst.SessionComponent).(*cherrySession.Component)
	if found == false {
		panic("session handlerComponent not found.")
	}
}

func (h *Handler) OnInit() {
}

func (h *Handler) OnAfterInit() {
}

func (h *Handler) OnStop() {
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

func (h *Handler) Component() *Component {
	return h.handlerComponent
}

func (h *Handler) AddLocals(localFns ...interface{}) {
	for _, fn := range localFns {
		funcName := cherryReflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.AddLocal(funcName, fn)
	}
}

func (h *Handler) AddLocal(name string, fn interface{}) {
	f, err := cherryReflect.GetInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	h.localHandlers[name] = f
}

func (h *Handler) AddRemotes(remoteFns ...interface{}) {
	for _, fn := range remoteFns {
		funcName := cherryReflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.AddRemote(funcName, fn)
	}
}

func (h *Handler) AddRemote(name string, fn interface{}) {
	invokeFunc, err := cherryReflect.GetInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	h.remoteHandlers[name] = invokeFunc
}

//AddEvent
func (h *Handler) AddEvent(eventName string, fn facade.EventFn) {
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
}

func (h *Handler) PostEvent(e facade.IEvent) {
	if h.handlerComponent == nil {
		cherryLogger.Errorf("handler handlerComponent is nil. event[%s]", e)
		return
	}

	h.handlerComponent.PostEvent(e)
}

func (h *Handler) AddOnCreate(listener ...cherrySession.SessionListener) {
	h.sessionComponent.AddOnCreate(listener...)
}

func (h *Handler) AddOnClose(listener ...cherrySession.SessionListener) {
	h.sessionComponent.AddOnClose(listener...)
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
