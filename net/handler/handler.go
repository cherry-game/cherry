package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
)

type (
	Handler struct {
		facade.AppContext
		name           string                       // unique name
		eventFn        map[string][]facade.EventFn  // event func
		localHandlers  map[string]*facade.HandlerFn // local invoke Handler functions
		remoteHandlers map[string]*facade.HandlerFn // remote invoke Handler functions
		component      *Component                   // handler component
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

	h.component = h.App().Find(cherryConst.HandlerComponent).(*Component)
	if h.component == nil {
		cherryLogger.Warn("not found component.")
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
	return h.component
}

func (h *Handler) RegisterLocals(localFns ...interface{}) {
	for _, fn := range localFns {
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
}

func (h *Handler) RegisterRemotes(remoteFns ...interface{}) {
	for _, fn := range remoteFns {
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
}

func (h *Handler) PostEvent(e facade.IEvent) {
	h.component.PostEvent(e)
}
