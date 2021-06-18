package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
	"runtime"
	"strings"
	"time"
)

type (
	//Component handler component
	Component struct {
		facade.Component // base component
		Options          // opts
		handlerGroups    []*HandlerGroup
		//rpc client
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
		handlerGroups: make([]*HandlerGroup, 0),
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
	for _, group := range h.handlerGroups {
		for _, handler := range group.handlers {
			handler.Set(h.App())
			handler.OnPreInit()
			handler.OnInit()
			handler.OnAfterInit()
		}

		for i := 0; i < group.queueNum; i++ {
			queue := group.queueMaps[i]

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

func (h *Component) invokeExecutor(executor IExecutor) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("recover in runQueue(). %s", r)
			var buf [512]byte
			n := runtime.Stack(buf[:], false)
			cherryLogger.Warnf("%s", string(buf[:n]))
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
		time.Sleep(3 * time.Second)
	}

	for _, group := range h.handlerGroups {
		for _, handler := range group.handlers {
			handler.OnStop()
		}
	}
}

func (h *Component) queueIsEmpty() bool {
	for _, group := range h.handlerGroups {
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
		cherryLogger.Warn("HandlerGroup is nil")
		return
	}

	for handlerName, handler := range handlerGroup.handlers {
		name := h.nameFn(handlerName)

		if name != handlerName {
			delete(handlerGroup.handlers, handlerName)
			handlerGroup.handlers[name] = handler
		}
	}

	// append to group
	h.handlerGroups = append(h.handlerGroups, handlerGroup)
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

	for _, group := range h.handlerGroups {
		for _, handler := range group.handlers {
			if fn, found := handler.Event(event.EventName()); found {
				executor := &EventExecutor{
					Event:   event,
					EventFn: fn,
				}

				index := 0
				if group.eventHash != nil {
					index = group.eventHash(executor, group.queueNum)
				} else {
					index = executor.HashQueue(group.queueNum)
				}
				group.inQueue(index, executor)
			}
		}
	}
}

func (h *Component) getGroup(handlerName string) (*HandlerGroup, facade.IHandler) {
	for _, group := range h.handlerGroups {
		if handler, found := group.handlers[handlerName]; found {
			return group, handler
		}
	}
	return nil, nil
}

func (h *Component) PostMessage(session *cherrySession.Session, route *cherryRoute.Route, msg *cherryMessage.Message) {
	if !h.App().Running() {
		//ignore message
		return
	}

	if route == nil || msg == nil {
		return
	}

	if route.NodeType() != h.App().NodeType() {
		//forward to remote server
		h.doForward(session, route, msg)
		return
	}

	handlerName := h.nameFn(route.HandleName())
	if handlerName == "" {
		cherryLogger.Warnf("could not find handle name. Route = %v", route)
		return
	}

	group, handler := h.getGroup(handlerName)
	if group == nil || handler == nil {
		cherryLogger.Warnf("[Route = %h] could not find handler for route.", msg.Route)
		return
	}

	fn, found := handler.LocalHandler(route.Method())
	if found == false {
		return
	}

	executor := &MessageExecutor{
		Session:       session,
		Msg:           msg,
		HandlerFn:     fn,
		BeforeFilters: h.beforeFilters,
		AfterFilters:  h.afterFilters,
	}

	index := 0
	if group.eventHash != nil {
		index = group.msgHash(executor, group.queueNum)
	} else {
		index = executor.HashQueue(group.queueNum)
	}
	group.inQueue(index, executor)
}

func (h *Component) doForward(session *cherrySession.Session, route *cherryRoute.Route, msg *cherryMessage.Message) {
	// TODO 通过rpc 转发到远程节点
	// rpc client invoke
	cherryLogger.Debugf("forward message. session[%s], route[%s], message[%s]", session, route, msg)
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
