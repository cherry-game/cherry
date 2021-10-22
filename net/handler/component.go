package cherryHandler

import (
	"context"
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryContext "github.com/cherry-game/cherry/net/context"
	"github.com/cherry-game/cherry/net/message"
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

	FilterFn func(ctx context.Context, session *cherrySession.Session, message *cherryMessage.Message) bool
)

func NewComponent(opts ...Option) *Component {
	component := &Component{
		groups: make([]*HandlerGroup, 0),
		options: options{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn:        strings.ToLower,
			printRouteLog: false,
		},
	}

	for _, opt := range opts {
		opt(&component.options)
	}

	return component
}

func (c *Component) Name() string {
	return cherryConst.HandlerComponent
}

func (c *Component) Init() {
}

func (c *Component) OnAfterInit() {
	//run handler group
	for _, g := range c.groups {
		g.run(c.App())
	}
}

func (c *Component) OnStop() {
	waitSecond := time.Duration(2)

	for {
		if c.queueIsEmpty() {
			for _, group := range c.groups {
				for _, handler := range group.handlers {
					handler.OnStop()
				}
			}
			return
		}

		// wait...
		cherryLogger.Debugf("queue is not empty! waiting %d seconds.", waitSecond.Seconds())
		time.Sleep(waitSecond)
	}
}

func (c *Component) queueIsEmpty() bool {
	for _, group := range c.groups {
		for _, queue := range group.queueMaps {
			if len(queue.dataChan) > 0 {
				return false
			}
		}
	}

	return true
}

func (c *Component) Register(handlerGroup *HandlerGroup) {
	if handlerGroup == nil {
		cherryLogger.Warn("handlerGroup is nil")
		return
	}

	for handlerName, handler := range handlerGroup.handlers {
		// process name fn
		name := c.nameFn(handlerName)

		if name != handlerName {
			delete(handlerGroup.handlers, handlerName)
			handlerGroup.handlers[name] = handler
		}
	}

	// append to group
	c.groups = append(c.groups, handlerGroup)
}

func (c *Component) Register2Group(handler ...facade.IHandler) {
	g := NewGroupWithHandler(handler...)
	c.Register(g)
}

// PostEvent 发布事件
func (c *Component) PostEvent(event facade.IEvent) {
	if event == nil {
		return
	}

	for _, group := range c.groups {
		for _, handler := range group.handlers {
			if eventSlice, found := handler.Event(event.Name()); found {
				executor := &ExecutorEvent{
					Event:      event,
					EventSlice: eventSlice,
				}

				index := group.queueHash(executor, group.queueNum)
				group.inQueue(index, executor)
			}
		}
	}
}

func (c *Component) GetHandler(route string) (*cherryMessage.Route, *HandlerGroup, facade.IHandler, bool) {
	r, err := cherryMessage.DecodeRoute(route)
	if err != nil {
		cherryLogger.Warnf("[Route = %s] decode fail.", route)
		return nil, nil, nil, false
	}

	handlerName := c.nameFn(r.HandleName())
	if handlerName == "" {
		cherryLogger.Warnf("[Route = %s] could not find handle name.", route)
		return nil, nil, nil, false
	}

	group, handler := c.getGroup(handlerName)
	if group == nil || handler == nil {
		cherryLogger.Warnf("[Route = %s] could not find handler group.", route)
		return nil, nil, nil, false
	}

	return r, group, handler, true
}

func (c *Component) getGroup(handlerName string) (*HandlerGroup, facade.IHandler) {
	for _, group := range c.groups {
		if handler, found := group.handlers[handlerName]; found {
			return group, handler
		}
	}
	return nil, nil
}

func (c *Component) DoLocalMessage(session *cherrySession.Session, msg *cherryMessage.Message) {
	if msg.RouteInfo().NodeType() == c.App().NodeType() {
		c.ProcessLocal(session, msg)
	} else {

	}
}

func (c *Component) ProcessLocal(session *cherrySession.Session, msg *cherryMessage.Message) {
	if !c.App().Running() {
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

	if msg.RouteInfo() == nil {
		err := msg.ParseRoute()
		if err != nil {
			session.Warnf("route decode error. route[%s], error[%s]", msg.Route, err)
			return
		}
	}

	if msg.RouteInfo().NodeType() != c.App().NodeType() {
		session.Warnf("msg node type error. route = %s", msg.Route)
		return
	}

	ctx := cherryContext.Add(context.Background(), cherryConst.MessageIdKey, msg.ID)
	ctx = cherryContext.Add(ctx, cherryConst.RouteKey, msg.Route)

	for _, filter := range c.beforeFilters {
		if filter(ctx, session, msg) == false {
			return
		}
	}

	rt, group, handler, found := c.GetHandler(msg.Route)
	if found == false {
		cherryLogger.Warnf("route not found handler. [route = %s]", msg.Route)
		return
	}

	fn, found := handler.LocalHandler(rt.Method())
	if found == false {
		cherryLogger.Debugf("[Route = %v] could not find method[%s] for route.", msg.Route, rt.Method())
		return
	}

	executor := &ExecutorLocal{
		IApplication: c.App(),
		Session:      session,
		Msg:          msg,
		HandlerFn:    fn,
		Ctx:          ctx,
		AfterFilters: c.afterFilters,
	}

	index := group.queueHash(executor, group.queueNum)
	group.inQueue(index, executor)

	if c.printRouteLog {
		session.Debugf("[local handler] [route = %s], [group-index = %d]",
			msg.RouteInfo(),
			index,
		)
	}
}

func (c *Component) ProcessRemote(group *HandlerGroup, executor *ExecutorRemote) {
	if !c.App().Running() {
		return
	}

	index := group.queueHash(executor, group.queueNum)
	group.inQueue(index, executor)

	if c.printRouteLog {
		cherryLogger.Debugf("[remote handler] [route = %s], [group-index = %d]",
			executor.String(),
			index,
		)
	}
}

func (c *Component) AddBeforeFilter(beforeFilters ...FilterFn) {
	if len(beforeFilters) > 0 {
		c.beforeFilters = append(c.beforeFilters, beforeFilters...)
	}
}

func (c *Component) AddAfterFilter(afterFilters ...FilterFn) {
	if len(afterFilters) > 0 {
		c.afterFilters = append(c.afterFilters, afterFilters...)
	}
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
