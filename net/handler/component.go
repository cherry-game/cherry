package cherryHandler

import (
	"fmt"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
)

type (
	//Component handler component
	Component struct {
		facade.Component                            // base component
		Options                                     // opts
		handlers         map[string]facade.IHandler // key:handlerName, value: Handler
		//rpc client
	}

	Options struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
	}

	UnhandledMessage struct {
		Session *cherrySession.Session
		Route   *cherryRoute.Route
		Msg     *cherryMessage.Message
	}

	FilterFn func(msg *UnhandledMessage) bool
)

func (u *UnhandledMessage) String() string {
	return fmt.Sprintf("session[%s], route[%s], msg[%s] ", u.Session, u.Route, u.Msg)
}

func NewComponent() *Component {
	return &Component{
		handlers: make(map[string]facade.IHandler),
		Options: Options{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn: func(s string) string {
				return s
			},
		},
	}
}

func (h *Component) Name() string {
	return cherryConst.HandlerComponent
}

func (h *Component) Init() {
	for _, handler := range h.handlers {
		handler.Set(h.App())
		handler.OnPreInit()
		handler.OnInit()
		handler.OnAfterInit()
	}
}

func (h *Component) OnStop() {
	for _, handler := range h.handlers {
		handler.OnStop()
	}
}

func (h *Component) Registers(handlers ...facade.IHandler) {
	for _, handler := range handlers {
		name := handler.Name()
		if name == "" {
			name = cherryReflect.GetStructName(handler)
		}
		h.RegisterWithName(name, handler)
	}
}

func (h *Component) RegisterWithName(name string, handler facade.IHandler) {
	if name == "" {
		cherryLogger.Warnf("[Handler = %s] name is empty. skipped.", cherryReflect.GetStructName(handler))
		return
	}

	if handler == nil {
		cherryLogger.Warnf("[Handler = %s] is empty. skipped.", name)
		return
	}

	name = h.nameFn(name)
	if name == "" {
		cherryLogger.Warnf("[Handler = %s] name is empty. skipped.", cherryReflect.GetStructName(handler))
		return
	}

	handler.SetName(name)

	if _, found := h.handlers[name]; found {
		cherryLogger.Errorf("[Handler = %s] is duplicate!", handler.Name())
		return
	}

	h.handlers[name] = handler
}

func (h *Component) DoHandle(msg *UnhandledMessage) {
	if !h.App().Running() {
		//ignore message
		return
	}

	if msg == nil || msg.Route == nil {
		return
	}

	if msg.Route.NodeType() != h.App().NodeType() {
		//forward to remote server
		h.doForward(msg)
		return
	} else {
		//TODO 消息过滤器

		handler := h.GetHandler(msg.Route)
		if handler == nil {
			cherryLogger.Errorf("[Route = %h] not found handler.", msg.Route)
			return
		}

		handler.PostMessage(msg)
	}
}

func (h *Component) doForward(msg *UnhandledMessage) {
	// TODO 通过rpc 转发到远程节点
	// rpc client invoke
	cherryLogger.Debugf("forward message. %s", msg)
}

func (h *Component) GetHandler(route *cherryRoute.Route) facade.IHandler {
	handlerName := h.nameFn(route.HandleName())
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
func (h *Component) PostEvent(event facade.IEvent) {
	if event == nil {
		return
	}

	for _, handler := range h.handlers {
		if _, found := handler.Event(event.EventName()); found {
			handler.PostMessage(event)
		}
	}
}

func (c *Options) GetBeforeFilter() []FilterFn {
	return c.beforeFilters
}

func (c *Options) BeforeFilter(beforeFilters ...FilterFn) {
	if len(beforeFilters) < 1 {
		return
	}
	c.beforeFilters = append(c.beforeFilters, beforeFilters...)
}

func (c *Options) GetAfterFilter() []FilterFn {
	return c.afterFilters
}

func (c *Options) AfterFilter(afterFilters ...FilterFn) {
	if len(afterFilters) < 1 {
		return
	}
	c.afterFilters = append(c.afterFilters, afterFilters...)
}

func (c *Options) SetNameFn(fn func(string) string) {
	if fn == nil {
		return
	}
	c.nameFn = fn
}

//// NodeRoute  结点路由规则 nodeType:结点类型,routeFunc 路由规则
//func (*HandlerOptions) NodeRoute(nodeType string, routeFunc cherryFacade.RouteFunction) {
//	cherryLogger.Panic(nodeType, routeFunc)
//}
