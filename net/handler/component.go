package cherryHandler

import (
	"context"
	ccode "github.com/cherry-game/cherry/code"
	cconst "github.com/cherry-game/cherry/const"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	ccontext "github.com/cherry-game/cherry/net/context"
	cmessage "github.com/cherry-game/cherry/net/message"
	csession "github.com/cherry-game/cherry/net/session"
	"github.com/nats-io/nats.go"
	"strings"
)

type (
	//Component handler component
	Component struct {
		options
		cfacade.Component
		groups []*HandlerGroup
	}

	options struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
	}

	Option func(options *options)

	FilterFn func(ctx context.Context, session *csession.Session, message *cmessage.Message) bool
)

func NewComponent(opts ...Option) *Component {
	component := &Component{
		groups: make([]*HandlerGroup, 0),
		options: options{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn:        strings.ToLower,
		},
	}

	for _, opt := range opts {
		opt(&component.options)
	}

	return component
}

func (c *Component) Name() string {
	return cconst.HandlerComponent
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
	for _, group := range c.groups {
		if group == nil {
			continue
		}

		for _, handler := range group.handlers {
			if handler != nil {
				handler.OnStop()
			}
		}
	}
}

func (c *Component) Register(handlerGroup *HandlerGroup) {
	if handlerGroup == nil {
		clog.Warn("handlerGroup is nil")
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

func (c *Component) Register2Group(handler ...cfacade.IHandler) {
	g := NewGroupWithHandler(handler...)
	c.Register(g)
}

// PostEvent 发布事件
func (c *Component) PostEvent(event cfacade.IEvent) {
	if event == nil {
		return
	}

	for _, group := range c.groups {
		for _, handler := range group.handlers {
			if eventInfo, found := handler.Event(event.Name()); found {
				executor := &ExecutorEvent{
					event:      event,
					eventSlice: eventInfo.List,
				}
				group.InQueue(eventInfo.QueueHash, executor)
			}
		}
	}
}

func (c *Component) GetHandler(route string) (*cmessage.Route, *HandlerGroup, cfacade.IHandler, bool) {
	r, err := cmessage.DecodeRoute(route)
	if err != nil {
		clog.Warnf("[Route = %s] decode fail.", route)
		return nil, nil, nil, false
	}

	handlerName := c.nameFn(r.HandleName())
	if handlerName == "" {
		clog.Warnf("[Route = %s] could not find handle name.", route)
		return nil, nil, nil, false
	}

	group, handler := c.getGroup(handlerName)
	if group == nil || handler == nil {
		clog.Warnf("[Route = %s] could not find handler group.", route)
		return nil, nil, nil, false
	}

	return r, group, handler, true
}

func (c *Component) getGroup(handlerName string) (*HandlerGroup, cfacade.IHandler) {
	for _, group := range c.groups {
		if handler, found := group.handlers[handlerName]; found {
			return group, handler
		}
	}
	return nil, nil
}

func (c *Component) ProcessLocal(session *csession.Session, msg *cmessage.Message) {
	if !c.App().Running() {
		return
	}

	if session == nil {
		clog.Debug("[local] session is nil")
		return
	}

	if msg == nil {
		session.Warn("[local] message is nil")
		return
	}

	if msg.RouteInfo() == nil {
		err := msg.ParseRoute()
		if err != nil {
			session.Warnf("[local] route decode error. [route = %s, error = %s]", msg.Route, err)
			return
		}
	}

	if msg.RouteInfo().NodeType() != c.App().NodeType() {
		session.Warnf("[local] msg node type error. [route = %s]", msg.Route)
		return
	}

	ctx := ccontext.Add(context.Background(), cconst.MessageIdKey, msg.ID)

	rt, group, handler, found := c.GetHandler(msg.Route)
	if found == false {
		clog.Warnf("[local] route not found handler. [route = %s]", msg.Route)
		return
	}

	fn, found := handler.LocalHandler(rt.Method())
	if found == false {
		clog.Debugf("[local] not find route. [Route = %v, method = %s]", msg.Route, rt.Method())
		return
	}

	executor := &ExecutorLocal{
		IApplication:  c.App(),
		session:       session,
		msg:           msg,
		handlerFn:     fn,
		ctx:           ctx,
		beforeFilters: c.beforeFilters,
		afterFilters:  c.afterFilters,
	}
	group.InQueue(fn.QueueHash, executor)
}

func (c *Component) ProcessRemote(route string, data []byte, natsMsg *nats.Msg) int32 {
	if !c.App().Running() {
		return ccode.AppIsStop
	}

	rt, group, handler, found := c.GetHandler(route)
	if found == false {
		clog.Warnf("handler not found. [route = %s]", route)
		return ccode.RPCHandlerError
	}

	fn, found := handler.RemoteHandler(rt.Method())
	if found == false {
		clog.Debugf("could not find route method. [Route = %v, method = %s].", route, rt.Method())
		return ccode.RPCHandlerError
	}

	if data == nil {
		data = []byte{}
	}

	executor := &ExecutorRemote{
		IApplication: c.App(),
		handlerFn:    fn,
		rt:           rt,
		data:         data,
		natsMsg:      natsMsg,
	}

	group.InQueue(fn.QueueHash, executor)
	return ccode.OK
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
