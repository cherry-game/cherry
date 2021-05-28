package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/reflect"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
)

type (
	//handlerComponent Handler component
	HandlerComponent struct {
		cherryFacade.Component                                  // base component
		HandlerOptions                                          // opts
		handlers               map[string]cherryFacade.IHandler // key:handlerName, value: Handler
	}

	HandlerOptions struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
	}

	UnhandledMessage struct {
		Session cherryFacade.ISession
		Route   *cherryRoute.Route
		Msg     *cherryMessage.Message
	}

	FilterFn func(msg *UnhandledMessage) bool
)

func NewComponent() *HandlerComponent {
	return &HandlerComponent{
		handlers: make(map[string]cherryFacade.IHandler),
		HandlerOptions: HandlerOptions{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn: func(s string) string {
				return s
			},
		},
	}
}

func (h *HandlerComponent) Name() string {
	return cherryConst.HandlerComponent
}

func (h *HandlerComponent) Init() {
	for _, handler := range h.handlers {
		handler.Set(h.App())
		handler.OnPreInit()
		handler.OnInit()
		handler.OnAfterInit()
	}
}

func (h *HandlerComponent) OnStop() {
	for _, handler := range h.handlers {
		handler.OnStop()
	}
}

func (h *HandlerComponent) Registers(handlers ...cherryFacade.IHandler) {
	for _, handler := range handlers {
		name := handler.Name()
		if name == "" {
			name = cherryReflect.GetStructName(handler)
		}
		h.RegisterWithName(name, handler)
	}
}

func (h *HandlerComponent) RegisterWithName(name string, handler cherryFacade.IHandler) {
	if name == "" {
		cherryLogger.Warnf("[Handler= %h] name is empty. skipped.", cherryReflect.GetStructName(handler))
		return
	}

	if handler == nil {
		cherryLogger.Warnf("[Handler= %s] is empty. skipped.", name)
		return
	}

	name = h.nameFn(name)
	if name == "" {
		cherryLogger.Warnf("[Handler= %h] name is empty. skipped.", cherryReflect.GetStructName(handler))
		return
	}

	handler.SetName(name)

	if _, found := h.handlers[name]; found {
		cherryLogger.Errorf("[Handler name = %s] is duplicate!", handler.Name())
		return
	}

	h.handlers[name] = handler
}

func (h *HandlerComponent) DoHandle(msg *UnhandledMessage) {
	if !h.App().Running() {
		//ignore message
		return
	}

	if msg == nil || msg.Route == nil {
		return
	}

	if msg.Route.NodeType() != h.App().NodeType() {
		//forward to remote server
		return
	}

	handler := h.GetHandler(msg.Route)
	if handler == nil {
		cherryLogger.Errorf("[Route = %h] not found handler.", msg.Route)
		return
	}

	handler.PostMessage(msg)
}

func (h *HandlerComponent) GetHandler(route *cherryRoute.Route) cherryFacade.IHandler {
	handlerName := h.nameFn(route.HandlerName())
	if handlerName == "" {
		cherryLogger.Warnf("could not find handle name. Route = %v", route)
		return nil
	}

	handler := h.handlers[handlerName]
	if handler == nil {
		cherryLogger.Warnf("could not find handle worker for Route = %v", route)
		return nil
	}
	return handler
}

// PostEvent 发布事件
func (h *HandlerComponent) PostEvent(event cherryFacade.IEvent) {
	if event == nil {
		return
	}

	for _, handler := range h.handlers {
		if _, found := handler.Event(event.EventName()); found {
			handler.PostMessage(event)
		}
	}
}

func (c *HandlerOptions) GetBeforeFilter() []FilterFn {
	return c.beforeFilters
}

func (c *HandlerOptions) BeforeFilter(beforeFilters ...FilterFn) {
	if len(beforeFilters) < 1 {
		return
	}
	c.beforeFilters = append(c.beforeFilters, beforeFilters...)
}

func (c *HandlerOptions) GetAfterFilter() []FilterFn {
	return c.afterFilters
}

func (c *HandlerOptions) AfterFilter(afterFilters ...FilterFn) {
	if len(afterFilters) < 1 {
		return
	}
	c.afterFilters = append(c.afterFilters, afterFilters...)
}

func (c *HandlerOptions) SetNameFn(fn func(string) string) {
	if fn == nil {
		return
	}
	c.nameFn = fn
}

// NodeRoute  结点路由规则 nodeType:结点类型,routeFunc 路由规则
func (*HandlerOptions) NodeRoute(nodeType string, routeFunc cherryFacade.RouteFunction) {
	cherryLogger.Panic(nodeType, routeFunc)
}
