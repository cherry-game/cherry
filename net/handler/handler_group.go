package cherryHandler

import (
	crypto "github.com/cherry-game/cherry/extend/crypto"
	"github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"runtime/debug"
)

type (
	HandlerGroup struct {
		//handlers  map[string]facade.IHandler
		handlers  map[string]facade.IHandler
		queueNum  int
		queueCap  int
		queueMaps map[int]*Queue
		queueHash QueueHash
	}

	Queue struct {
		index    int
		dataChan chan IExecutor
	}

	QueueHash func(executor IExecutor, queueNum int) int
)

func NewGroupWithHandler(handlers ...facade.IHandler) *HandlerGroup {
	g := NewGroup(1, 128)
	g.AddHandlers(handlers...)
	return g
}

func NewGroup(queueNum, queueCap int) *HandlerGroup {
	if queueNum < 1 || queueNum > 32767 {
		queueNum = 1
	}

	if queueCap < 1 || queueCap > 32767 {
		queueCap = 128
	}

	g := &HandlerGroup{
		handlers:  make(map[string]facade.IHandler),
		queueNum:  queueNum,
		queueCap:  queueCap,
		queueMaps: make(map[int]*Queue),
		queueHash: DefaultQueueHash, // default queue hash
	}

	// init queue chan
	for i := 0; i < queueNum; i++ {
		q := &Queue{
			index:    i,
			dataChan: make(chan IExecutor, queueCap),
		}
		g.queueMaps[i] = q
	}

	return g
}

func (h *HandlerGroup) AddHandlers(handlers ...facade.IHandler) {
	for _, handler := range handlers {
		if handler.Name() == "" {
			handler.SetName(cherryReflect.GetStructName(handler))
		}

		h.handlers[handler.Name()] = handler
	}
}

func (h *HandlerGroup) SetQueueHash(fn QueueHash) {
	h.queueHash = fn
}

func (h *HandlerGroup) inQueue(index int, executor IExecutor) {
	if index > h.queueNum {
		index = 0
	}

	q := h.queueMaps[index]
	q.dataChan <- executor
}

func (h *HandlerGroup) run(app facade.IApplication) {
	for _, handler := range h.handlers {
		handler.Set(app)
		handler.OnPreInit()
		handler.OnInit()
		handler.OnAfterInit()
		h.printInfo(handler)
	}

	for i := 0; i < h.queueNum; i++ {
		queue := h.queueMaps[i]
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

func (h *HandlerGroup) invokeExecutor(executor IExecutor) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("recover in executor. %s", string(debug.Stack()))
			cherryLogger.Warnf("executor fail [%s]", executor.String())
		}
	}()

	executor.Invoke()
}

func (h *HandlerGroup) printInfo(handler facade.IHandler) {
	cherryLogger.Debug("--------------------------------------")
	cherryLogger.Debugf("[Handler = %s] queueNum = %d, queueCap = %d", handler.Name(), h.queueNum, h.queueCap)
	for key := range handler.Events() {
		cherryLogger.Debugf("[Handler = %s] event = %s", handler.Name(), key)
	}

	for key := range handler.LocalHandlers() {
		cherryLogger.Debugf("[Handler = %s] localHandler = %s", handler.Name(), key)
	}

	for key := range handler.RemoteHandlers() {
		cherryLogger.Debugf("[Handler = %s] removeHandler = %s", handler.Name(), key)
	}
	cherryLogger.Debug("--------------------------------------")
}

func DefaultQueueHash(executor IExecutor, queueNum int) int {
	var i = 0
	switch e := executor.(type) {
	case *MessageExecutor:
		if e.Session.UID() > 0 {
			i = int(e.Session.UID() % int64(queueNum))
		} else {
			i = int(e.Session.SID() % int64(queueNum))
		}
	case *EventExecutor:
		i = crypto.CRC32(e.Event.UniqueId()) % queueNum
	case *UserRPCExecutor:
		i = 0

	case *SysRPCExecutor:
		i = 0
	}

	return i
}
