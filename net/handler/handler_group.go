package cherryHandler

import (
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"runtime/debug"
)

type (
	HandlerGroup struct {
		handlers  map[string]cfacade.IHandler
		queueNum  int
		queueCap  int
		queueMaps map[int]*Queue
		queueHash cfacade.QueueHashFn
	}

	Queue struct {
		index    int
		dataChan chan cfacade.IExecutor
	}
)

func NewGroupWithHandler(handlers ...cfacade.IHandler) *HandlerGroup {
	g := NewGroup(1, 512)
	g.AddHandlers(handlers...)
	return g
}

func NewGroup(queueNum, queueCap int) *HandlerGroup {
	if queueNum < 1 || queueNum > 32767 {
		queueNum = 1
	}

	if queueCap < 1 || queueCap > 32767 {
		queueCap = 512
	}

	g := &HandlerGroup{
		handlers:  make(map[string]cfacade.IHandler),
		queueNum:  queueNum,
		queueCap:  queueCap,
		queueMaps: make(map[int]*Queue),
	}

	// init queue chan
	for i := 0; i < queueNum; i++ {
		q := &Queue{
			index:    i,
			dataChan: make(chan cfacade.IExecutor, queueCap),
		}
		g.queueMaps[i] = q
	}

	return g
}

func (h *HandlerGroup) AddHandlers(handlers ...cfacade.IHandler) {
	for _, handler := range handlers {
		if handler.Name() == "" {
			handler.SetName(creflect.GetStructName(handler))
		}

		h.handlers[handler.Name()] = handler
	}
}

func (h *HandlerGroup) SetQueueHash(fn cfacade.QueueHashFn) {
	h.queueHash = fn
}

func (h *HandlerGroup) InQueue(hashFn cfacade.QueueHashFn, executor cfacade.IExecutor) {
	index := 0

	if h.queueNum > 1 {
		if hashFn != nil {
			index = hashFn(executor, h.queueNum)
		} else if h.queueHash != nil {
			index = h.queueHash(executor, h.queueNum)
		} else {
			index = executor.QueueHash(h.queueNum)
		}
	}

	if index > h.queueNum {
		clog.Errorf("group index error. [groupIndex = %d, queueNum = %d, executor = %v]",
			index,
			h.queueNum,
			executor,
		)
		return
	}

	executor.SetIndex(index)

	if q, found := h.queueMaps[index]; found {
		q.dataChan <- executor
	}
}

func (h *HandlerGroup) run(app cfacade.IApplication) {
	for _, handler := range h.handlers {
		handler.Set(app)
		handler.OnPreInit()
	}

	for _, handler := range h.handlers {
		handler.OnInit()
	}

	for _, handler := range h.handlers {
		handler.OnAfterInit()
		h.printInfo(handler)
	}

	for _, queue := range h.queueMaps {
		go queue.run()
	}
}

func (q *Queue) run() {
	for {
		select {
		case executor := <-q.dataChan:
			{
				q.executorInvoke(executor)
			}
		}
	}
}

func (q *Queue) executorInvoke(executor cfacade.IExecutor) {
	defer func() {
		if rev := recover(); rev != nil {
			clog.Warnf("recover in handle group. %s", string(debug.Stack()))
		}
	}()

	executor.Invoke()
}

func (h *HandlerGroup) printInfo(handler cfacade.IHandler) {
	clog.Infof("[handler = %s] queueNum = %d, queueCap = %d", handler.Name(), h.queueNum, h.queueCap)
	for key := range handler.Events() {
		clog.Infof("[handler = %s] event = %s", handler.Name(), key)
	}

	for key := range handler.LocalHandlers() {
		clog.Infof("[handler = %s] localMethod = %s", handler.Name(), key)
	}

	for key := range handler.RemoteHandlers() {
		clog.Infof("[handler = %s] remoteHandler = %s", handler.Name(), key)
	}
}
