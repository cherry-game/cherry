package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
	"strings"
	"time"
)

type (
	//Component handler component
	Component struct {
		options
		facade.Component
		groups []*HandlerGroup
	}

	options struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
		printRouteLog bool
	}

	Option func(options *options)

	FilterFn func(session *cherrySession.Session, message *cherryMessage.Message) bool
)

func NewComponent(opts ...Option) *Component {
	c := &Component{
		groups: make([]*HandlerGroup, 0),
		options: options{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn:        strings.ToLower,
			printRouteLog: false,
		},
	}

	for _, opt := range opts {
		opt(&c.options)
	}

	return c
}

func (h *Component) Name() string {
	return cherryConst.HandlerComponent
}

func (h *Component) Init() {
	//run handler group
	for _, g := range h.groups {
		g.run(h.App())
	}
}

func (h *Component) OnStop() {
	waitSecond := time.Duration(2)

	for {
		if h.queueIsEmpty() {
			for _, group := range h.groups {
				for _, handler := range group.handlers {
					handler.OnStop()
				}
			}
			return
		}

		// wait...
		cherryLogger.Debugf("queue is not empty! wait %d seconds.", waitSecond.Seconds())
		time.Sleep(waitSecond)
	}
}

func (h *Component) queueIsEmpty() bool {
	for _, group := range h.groups {
		for _, queue := range group.queueMaps {
			if len(queue.dataChan) > 0 {
				return false
			}
		}
	}

	return true
}

func (h *Component) Register(handlerGroup *HandlerGroup) {
	if handlerGroup == nil {
		cherryLogger.Warn("handlerGroup is nil")
		return
	}

	for handlerName, handler := range handlerGroup.handlers {
		// process name fn
		name := h.nameFn(handlerName)

		if name != handlerName {
			delete(handlerGroup.handlers, handlerName)
			handlerGroup.handlers[name] = handler
		}
	}

	// append to group
	h.groups = append(h.groups, handlerGroup)
}

func (h *Component) Register2Group(handler ...facade.IHandler) {
	g := NewGroupWithHandler(handler...)
	h.Register(g)
}

// PostEvent 发布事件
func (h *Component) PostEvent(event facade.IEvent) {
	if event == nil {
		return
	}

	for _, group := range h.groups {
		for _, handler := range group.handlers {

			if eventSlice, found := handler.Event(event.Name()); found {
				executor := &EventExecutor{
					Event:      event,
					EventSlice: eventSlice,
				}

				index := group.queueHash(executor, group.queueNum)
				group.inQueue(index, executor)
			}
		}
	}
}

func (h *Component) getGroup(handlerName string) (*HandlerGroup, facade.IHandler) {
	for _, group := range h.groups {
		if handler, found := group.handlers[handlerName]; found {
			return group, handler
		}
	}
	return nil, nil
}

func (h *Component) PostMessage(session *cherrySession.Session, msg *cherryMessage.Message) {
	if !h.App().Running() {
		//ignore message
		return
	}

	if session == nil {
		cherryLogger.Debug("session is nil")
		return
	}

	if msg == nil {
		session.Warn("message is nil")
		return
	}

	route, err := cherryRoute.Decode(msg.Route)
	if err != nil {
		session.Warnf("route decode error. route[%s], error[%s]", msg.Route, err)
		return
	}

	if route.NodeType() != h.App().NodeType() {
		//forward to remote server
		h.forwardToRemote(session, msg)
		return
	}

	handlerName := h.nameFn(route.HandleName())
	if handlerName == "" {
		cherryLogger.Warnf("[Route = %v] could not find handle name. ", route)
		return
	}

	group, handler := h.getGroup(handlerName)
	if group == nil || handler == nil {
		cherryLogger.Warnf("[Route = %v] could not find handler for route.", route)
		return
	}

	fn, found := handler.LocalHandler(route.Method())
	if found == false {
		cherryLogger.Debugf("[Route = %v] could not find method[%s] for route.", route.Method())
		return
	}

	executor := &MessageExecutor{
		App:           h.App(),
		Session:       session,
		Msg:           msg,
		HandlerFn:     fn,
		BeforeFilters: h.beforeFilters,
		AfterFilters:  h.afterFilters,
	}

	index := group.queueHash(executor, group.queueNum)

	if h.printRouteLog {
		session.Debugf("post message handler[%s], route[%s], group-index[%d]", handlerName, route, index)
	}

	group.inQueue(index, executor)
}

func (h *Component) AddBeforeFilter(beforeFilters ...FilterFn) {
	if len(beforeFilters) > 0 {
		h.beforeFilters = append(h.beforeFilters, beforeFilters...)
	}
}

func (h *Component) AddAfterFilter(afterFilters ...FilterFn) {
	if len(afterFilters) > 0 {
		h.afterFilters = append(h.afterFilters, afterFilters...)
	}
}

func (h *Component) forwardToRemote(session *cherrySession.Session, msg *cherryMessage.Message) {
	// TODO 通过rpc 转发到远程节点
	cherryLogger.Warnf("forward to remote session[%s], msg[%s]", session, msg)
}

func WithBeforeFilter(beforeFilters ...FilterFn) Option {
	return func(options *options) {
		if len(beforeFilters) > 0 {
			options.beforeFilters = append(options.beforeFilters, beforeFilters...)
		}
	}
}

func WithAfterFilter(afterFilters ...FilterFn) Option {
	return func(options *options) {
		if len(afterFilters) > 0 {
			options.afterFilters = append(options.afterFilters, afterFilters...)
		}
	}
}

func WithNameFunc(fn func(string) string) Option {
	return func(options *options) {
		if fn != nil {
			options.nameFn = fn
		}
	}
}

func WithPrintRouteLog(enable bool) Option {
	return func(options *options) {
		options.printRouteLog = enable
	}
}
