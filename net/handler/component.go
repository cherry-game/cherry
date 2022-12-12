package cherryHandler

import (
	"context"
	ccode "github.com/cherry-game/cherry/code"
	cconst "github.com/cherry-game/cherry/const"
	cherryQueue "github.com/cherry-game/cherry/extend/queue"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	ccontext "github.com/cherry-game/cherry/net/context"
	cmsg "github.com/cherry-game/cherry/net/message"
	csession "github.com/cherry-game/cherry/net/session"
	"github.com/nats-io/nats.go"
	"strings"
	"time"
)

const (
	Name = "handler_component"
)

type (
	//Component handler component
	Component struct {
		options
		cfacade.Component
		closeChan  chan bool
		groups     []*HandlerGroup
		eventQueue *cherryQueue.Queue
	}

	options struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
	}

	Option func(options *options)

	FilterFn func(ctx context.Context, session *csession.Session, message *cmsg.Message) bool
)

func NewComponent(opts ...Option) *Component {
	component := &Component{
		groups: make([]*HandlerGroup, 0),
		options: options{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn:        strings.ToLower,
		},
		closeChan:  make(chan bool, 0),
		eventQueue: cherryQueue.NewQueue(),
	}

	for _, opt := range opts {
		opt(&component.options)
	}

	return component
}

func (c *Component) Name() string {
	return Name
}

func (c *Component) Init() {
	go c.runEventChan()
}

func (c *Component) postEventToQueue(num int) {
	for i := 0; i < num; i++ {
		e, ok := c.eventQueue.Pop()
		if !ok {
			return
		}

		event, ok := e.(cfacade.IEvent)
		if !ok {
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
}

func (c *Component) runEventChan() {
	postTicker := time.NewTicker(10 * time.Millisecond)
	postNum := 1000

	for {
		select {
		case <-postTicker.C:
			{
				c.postEventToQueue(postNum)
			}
		case <-c.closeChan:
			{
				postTicker.Stop()
				clog.Infof("execute component close chan.")
				break
			}
		}
	}
}

func (c *Component) OnAfterInit() {
	//run handler group
	for _, g := range c.groups {
		g.run(c.App())
	}
}

func (c *Component) OnStop() {
	c.closeChan <- true

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

	c.eventQueue.Push(event)
	//c.eventChan <- event
}

func (c *Component) GetHandler(handlerName string) (*HandlerGroup, cfacade.IHandler, bool) {
	handlerName = c.nameFn(handlerName)
	if handlerName == "" {
		clog.Warnf("[handlerName = %s] could not find handle name.", handlerName)
		return nil, nil, false
	}

	group, handler := c.getGroup(handlerName)
	if group == nil || handler == nil {
		clog.Warnf("[handlerName = %s] could not find handler group.", handlerName)
		return nil, nil, false
	}

	return group, handler, true
}

func (c *Component) getGroup(handlerName string) (*HandlerGroup, cfacade.IHandler) {
	for _, group := range c.groups {
		if handler, found := group.handlers[handlerName]; found {
			return group, handler
		}
	}
	return nil, nil
}

func (c *Component) ProcessLocal(session *csession.Session, msg *cmsg.Message) {
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
		if err := msg.ParseRoute(); err != nil {
			session.Warnf("[local] route decode error. [route = %s, error = %s]", msg.Route, err)
			return
		}
	}

	routeInfo := msg.RouteInfo()
	group, handler, found := c.GetHandler(routeInfo.HandleName())
	if found == false {
		clog.Warnf("[local] route not found handler. [route = %s]", msg.Route)
		return
	}

	fn, found := handler.LocalHandler(routeInfo.Method())
	if found == false {
		clog.Debugf("[local] not find route. [Route = %v, method = %s]", msg.Route, routeInfo.Method())
		return
	}

	ctx := ccontext.Add(context.Background(), cconst.MessageIdKey, msg.ID)

	executor := &ExecutorLocal{
		IApplication:  c.App(),
		session:       session,
		msg:           msg,
		handlerFn:     fn,
		ctx:           ctx,
		beforeFilters: c.beforeFilters,
		afterFilters:  c.afterFilters,
	}

	// in queue
	group.InQueue(fn.QueueHash, executor)
}

func (c *Component) ProcessRemote(route string, data []byte, natsMsg *nats.Msg) int32 {
	if !c.App().Running() {
		return ccode.AppIsStop
	}

	routeInfo, err := cmsg.DecodeRoute(route)
	if err != nil {
		clog.Warnf("route decode error. [route = %s]", route)
		return ccode.RPCRouteDecodeError
	}

	group, handler, found := c.GetHandler(routeInfo.HandleName())
	if found == false {
		clog.Warnf("handler not found. [route = %s]", route)
		return ccode.RPCHandlerError
	}

	fn, found := handler.RemoteHandler(routeInfo.Method())
	if found == false {
		clog.Debugf("could not find route method. [Route = %v, method = %s].", route, routeInfo.Method())
		return ccode.RPCHandlerError
	}

	if data == nil {
		data = []byte{}
	}

	executor := &ExecutorRemote{
		IApplication: c.App(),
		handlerFn:    fn,
		rt:           routeInfo,
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

func WithName(fn func(string) string) Option {
	return func(options *options) {
		if fn != nil {
			options.nameFn = fn
		}
	}
}

func WithEventQueueCap(eventQueueCap int) Option {
	return func(options *options) {
		if eventQueueCap > 0 {
			//options.eventChan = make(chan cfacade.IEvent, eventQueueCap)
		}
	}
}
