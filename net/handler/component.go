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
	//Component handler handlerComponent
	Component struct {
		Options
		facade.Component
		groups []*HandlerGroup
	}

	Options struct {
		beforeFilters []FilterFn
		afterFilters  []FilterFn
		nameFn        func(string) string
		printRouteLog bool
	}

	FilterFn func(session *cherrySession.Session, message *cherryMessage.Message) bool
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
	//run handler group
	for _, g := range h.groups {
		g.run(h.App())
	}
}

func (h *Component) OnStop() {
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
		cherryLogger.Debug("queue not empty! wait 3 seconds.")
		time.Sleep(1 * time.Second)
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

// 本地handler执行， remote用于远程rpc调用执行
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

func (h *Component) forwardToRemote(session *cherrySession.Session, msg *cherryMessage.Message) {
	// TODO 通过rpc 转发到远程节点
	cherryLogger.Warnf("forward to remote session[%s], msg[%s]", session, msg)
}

func (c *Options) AddBeforeFilter(beforeFilters ...FilterFn) {
	if len(beforeFilters) > 0 {
		c.beforeFilters = append(c.beforeFilters, beforeFilters...)
	}
}

func (c *Options) AddAfterFilter(afterFilters ...FilterFn) {
	if len(afterFilters) > 0 {
		c.afterFilters = append(c.afterFilters, afterFilters...)
	}
}

func (c *Options) SetNameFn(fn func(string) string) {
	if fn == nil {
		return
	}
	c.nameFn = fn
}

func (c *Options) PrintRouteLog(enable bool) {
	c.printRouteLog = enable
}
