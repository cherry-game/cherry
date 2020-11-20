package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/utils"
	"reflect"
	"strings"
)

type (
	//handlerComponent Handler component
	HandlerComponent struct {
		cherryInterfaces.BaseComponent                             // base component
		HandlerComponentOptions                                    // opts
		iHandlers                      []cherryInterfaces.IHandler // register handlers
		workers                        map[string]*Worker          // wrap handler for worker
	}

	HandlerComponentOptions struct {
		beforeFilters []FilterFunc
		afterFilters  []FilterFunc
		nameFunc      func(string) string
	}

	UnhandledMessage struct {
		Session cherryInterfaces.ISession
		Route   *cherryNet.Route
		Msg     *cherryNetMessage.Message
	}

	FilterFunc func(msg UnhandledMessage) bool
)

func NewComponent() *HandlerComponent {
	return &HandlerComponent{
		workers: make(map[string]*Worker),
		HandlerComponentOptions: HandlerComponentOptions{
			beforeFilters: make([]FilterFunc, 0),
			afterFilters:  make([]FilterFunc, 0),
			nameFunc:      strings.ToLower,
		},
	}
}

func (h *HandlerComponent) Name() string {
	return cherryConst.HandlerComponent
}

func (h *HandlerComponent) Init() {

	for _, iHandler := range h.iHandlers {
		iHandler.Set(h.App())
		iHandler.Init()

		handler := convert2Handler(iHandler)
		if handler == nil {
			cherryLogger.Warnf("IHandler convert to Handler fail. name=%s", iHandler.Name())
			return
		}

		handlerName := h.nameFunc(handler.name)

		if _, found := h.workers[handlerName]; found {
			cherryLogger.Errorf("[Handler name = %s] is duplicate!", handler.name)
			break
		}

		handler.Init()

		worker := NewWorker(h, handler)
		worker.Start()

		h.workers[handlerName] = worker
	}

	cherryLogger.Debug("[handlerComponent] init completed.")
}

func (h *HandlerComponent) Stop() {
}

func (h *HandlerComponent) Register(handlers ...cherryInterfaces.IHandler) {
	for _, handler := range handlers {
		name := handler.Name()
		if name == "" {
			name = cherryUtils.Reflect.GetStructName(handler)
		}
		h.RegisterWithName(name, handler)
	}
}

func (h *HandlerComponent) RegisterWithName(name string, handler cherryInterfaces.IHandler) {
	if name == "" {
		cherryLogger.Warnf("[Handler= %h] name is empty. skipped.", cherryUtils.Reflect.GetStructName(handler))
		return
	}

	if handler == nil {
		cherryLogger.Warnf("[Handler= %s] is empty. skipped.", name)
		return
	}

	handler.SetName(name)

	h.iHandlers = append(h.iHandlers, handler)
}

func convert2Handler(handler cherryInterfaces.IHandler) *Handler {
	t := reflect.TypeOf(handler)
	v := reflect.ValueOf(handler)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		if v.Field(i).Type() == reflect.TypeOf(Handler{}) {
			base := v.Field(i).Interface().(Handler)
			return &base
		}
	}
	return nil
}

func (h *HandlerComponent) InHandle(route *cherryNet.Route, session cherryInterfaces.ISession, message *cherryNetMessage.Message) {
	if route.NodeType() != h.App().NodeType() {
		return
	}

	if !h.App().Running() {
		//ignore IMessage
		return
	}

	worker := h.GetWorker(route)
	if worker == nil {
		cherryLogger.Errorf("[Route = %h] not found handle worker.", route)
		return
	}

	msg := UnhandledMessage{
		Session: session,
		Route:   route,
		Msg:     message,
	}

	worker.PutMessage(msg)
}

func (h *HandlerComponent) GetWorker(route *cherryNet.Route) *Worker {
	worker := h.workers[h.nameFunc(route.HandlerName())]
	if worker == nil {
		cherryLogger.Warnf("could not find handle worker for Route = %h", route)
		return nil
	}
	return worker
}

// PostEvent 发布事件
func (h *HandlerComponent) PostEvent(event cherryInterfaces.IEvent) {
	if event == nil {
		return
	}

	for _, worker := range h.workers {
		if _, found := worker.GetEvent(event.EventName()); found {
			worker.PutMessage(event)
		}
	}
}

func (c *HandlerComponentOptions) GetBeforeFilter() []FilterFunc {
	return c.beforeFilters
}

func (c *HandlerComponentOptions) BeforeFilter(beforeFilters ...FilterFunc) {
	if len(beforeFilters) < 1 {
		return
	}
	c.beforeFilters = append(c.beforeFilters, beforeFilters...)
}

func (c *HandlerComponentOptions) GetAfterFilter() []FilterFunc {
	return c.afterFilters
}

func (c *HandlerComponentOptions) AfterFilter(afterFilters ...FilterFunc) {
	if len(afterFilters) < 1 {
		return
	}
	c.afterFilters = append(c.afterFilters, afterFilters...)
}

func (c *HandlerComponentOptions) SetNameFunc(fn func(string) string) {
	if fn == nil {
		return
	}
	c.nameFunc = fn
}

// NodeRoute  结点路由规则 nodeType:结点类型,routeFunc 路由规则
func (*HandlerComponentOptions) NodeRoute(nodeType string, routeFunc cherryInterfaces.RouteFunction) {
	cherryLogger.Panic(nodeType, routeFunc)
}
