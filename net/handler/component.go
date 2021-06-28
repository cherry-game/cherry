package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"runtime/debug"
	"strings"
	"time"
)

type (
	//Component handler component
	Component struct {
		Options
		facade.Component
		groups []*HandlerGroup
	}

	Options struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
	}

	FilterFn func(msg *MessageExecutor) bool
)

func NewComponent() *Component {
	return &Component{
		groups: make([]*HandlerGroup, 0),
		Options: Options{
			beforeFilters: make([]FilterFn, 0),
			afterFilters:  make([]FilterFn, 0),
			nameFn:        strings.ToLower,
		},
	}
}

func (h *Component) Name() string {
	return cherryConst.HandlerComponent
}

func (h *Component) Init() {
	for _, g := range h.groups {
		for _, handler := range g.handlers {
			handler.Set(h.App())
			handler.OnPreInit()
			handler.OnInit()
			handler.OnAfterInit()

			printHandler(g, handler)
		}

		for i := 0; i < g.queueNum; i++ {
			queue := g.queueMaps[i]

			// new goroutine for queue
			go func(queue *Queue) {
				for {
					select {
					case executor := <-queue.dataChan:
						{
							h.invokeExecutor(executor)
						}
					}
				}
			}(queue)
		}
	}
}

func printHandler(g *HandlerGroup, handler facade.IHandler) {
	cherryLogger.Debugf("[Handler = %s] queueNum = %d, queueCap = %d", handler.Name(), g.queueNum, g.queueCap)

	for key := range handler.Events() {
		cherryLogger.Debugf("[Handler = %s] event = %s", handler.Name(), key)
	}

	for key := range handler.LocalHandlers() {
		cherryLogger.Debugf("[Handler = %s] localHandler = %s", handler.Name(), key)
	}

	for key := range handler.RemoteHandlers() {
		cherryLogger.Debugf("[Handler = %s] removeHandler = %s", handler.Name(), key)
	}
}

func (h *Component) invokeExecutor(executor IExecutor) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("recover in runQueue(). %s", string(debug.Stack()))
		}
	}()

	executor.Invoke()
}

func (h *Component) OnStop() {
	for {
		if h.queueIsEmpty() {
			return
		}
		// wait...
		cherryLogger.Debug("queue not empty! wait 3 seconds.")
		time.Sleep(1 * time.Second)
	}

	for _, group := range h.groups {
		for _, handler := range group.handlers {
			handler.OnStop()
		}
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

			if fn, found := handler.Event(event.EventName()); found {
				executor := &EventExecutor{
					Event:   event,
					EventFn: fn,
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

func (h *Component) PostMessage(agent *cherryAgent.Agent, route *cherryRoute.Route, msg *cherryMessage.Message) {
	if !h.App().Running() {
		//ignore message
		return
	}

	if route == nil {
		cherryLogger.Debug("route is nil")
		return
	}

	if msg == nil {
		cherryLogger.Debug("data is nil")
		return
	}

	if route.NodeType() != h.App().NodeType() {
		//forward to remote server
		h.doForward(agent.Session, route, msg)
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
		Agent:         agent,
		Msg:           msg,
		HandlerFn:     fn,
		BeforeFilters: h.beforeFilters,
		AfterFilters:  h.afterFilters,
	}

	index := group.queueHash(executor, group.queueNum)

	if cherryProfile.Debug() {
		//agent.Session.Debugf("post handler = %s, route = %s, group-index = %d", handlerName, route, index)
	}

	group.inQueue(index, executor)
}

func (h *Component) doForward(session *cherrySession.Session, route *cherryRoute.Route, msg *cherryMessage.Message) {
	// TODO 通过rpc 转发到远程节点
	// rpc client invoke

	session.Debugf("forward message. route[%s], message[%s]", route, msg)
}

func (c *Options) AddBeforeFilter(beforeFilters ...FilterFn) {
	if len(beforeFilters) < 1 {
		return
	}
	c.beforeFilters = append(c.beforeFilters, beforeFilters...)
}

func (c *Options) AddAfterFilter(afterFilters ...FilterFn) {
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
