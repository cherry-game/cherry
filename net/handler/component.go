package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
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
		groups        []*HandlerGroup
		RemoteHandler cherryAgent.RPCHandler
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
		cherryLogger.Debugf("queue is not empty! wait %d seconds.", waitSecond.Seconds())
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

func (c *Component) getGroup(handlerName string) (*HandlerGroup, facade.IHandler) {
	for _, group := range c.groups {
		if handler, found := group.handlers[handlerName]; found {
			return group, handler
		}
	}
	return nil, nil
}

func (c *Component) PostMessage(session *cherrySession.Session, msg *cherryMessage.Message) {
	if !c.App().Running() {
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

	err := msg.ParseRoute()
	if err != nil {
		session.Warnf("route decode error. route[%s], error[%s]", msg.Route, err)
		return
	}

	if msg.RouteInfo().NodeType() == c.App().NodeType() {
		c.localProcess(session, msg)
	} else {
		c.remoteProcess(session, msg)
	}
}

func (c *Component) localProcess(session *cherrySession.Session, msg *cherryMessage.Message) {
	handlerName := c.nameFn(msg.RouteInfo().HandleName())
	if handlerName == "" {
		cherryLogger.Warnf("[Route = %v] could not find handle name. ", msg.RouteInfo())
		return
	}

	group, handler := c.getGroup(handlerName)
	if group == nil || handler == nil {
		cherryLogger.Warnf("[Route = %v] could not find handler for route.", msg.RouteInfo())
		return
	}

	fn, found := handler.LocalHandler(msg.RouteInfo().Method())
	if found == false {
		cherryLogger.Debugf("[Route = %v] could not find method[%s] for route.", msg.RouteInfo().Method())
		return
	}

	executor := &MessageExecutor{
		App:           c.App(),
		Session:       session,
		Msg:           msg,
		HandlerFn:     fn,
		BeforeFilters: c.beforeFilters,
		AfterFilters:  c.afterFilters,
	}

	index := group.queueHash(executor, group.queueNum)

	if c.printRouteLog {
		session.Debugf("execute handler[%s], route[%s], group-index[%d]", handlerName, msg.RouteInfo(), index)
	}

	group.inQueue(index, executor)
}

func (c *Component) remoteProcess(session *cherrySession.Session, msg *cherryMessage.Message) {
	if c.RemoteHandler != nil {
		c.RemoteHandler(session, msg)
	} else {
		c.localProcess(session, msg)
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
