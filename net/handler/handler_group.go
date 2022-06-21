package cherryHandler

import (
	crypto "github.com/cherry-game/cherry/extend/crypto"
	"github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"math/rand"
	"runtime/debug"
)

type (
	HandlerGroup struct {
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

func (h *HandlerGroup) InQueue(executor IExecutor) {
	index := h.queueHash(executor, h.queueNum)
	executor.SetIndex(index)

	if index > h.queueNum {
		cherryLogger.Errorf("group index error. [groupIndex = %d, queueNum = %d]", executor.Index(), h.queueNum)
		return
	}

	q := h.queueMaps[index]
	q.dataChan <- executor
}

func (h *HandlerGroup) run(app facade.IApplication) {
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

func (q *Queue) executorInvoke(executor IExecutor) {
	defer func() {
		if rev := recover(); rev != nil {
			cherryLogger.Warnf("recover in handle group. %s", string(debug.Stack()))
		}
	}()

	executor.Invoke()
}

func (h *HandlerGroup) printInfo(handler facade.IHandler) {
	cherryLogger.Infof("[handler = %s] queueNum = %d, queueCap = %d", handler.Name(), h.queueNum, h.queueCap)
	for key := range handler.Events() {
		cherryLogger.Infof("[handler = %s] event = %s", handler.Name(), key)
	}

	for key := range handler.LocalHandlers() {
		cherryLogger.Infof("[handler = %s] localHandler = %s", handler.Name(), key)
	}

	for key := range handler.RemoteHandlers() {
		cherryLogger.Infof("[handler = %s] remoteHandler = %s", handler.Name(), key)
	}
}

func DefaultQueueHash(executor IExecutor, queueNum int) int {
	if queueNum <= 1 {
		return 0
	}

	var i = 0
	switch e := executor.(type) {
	case *ExecutorLocal:
		if e.Session.UID() > 0 {
			i = int(e.Session.UID() % int64(queueNum))
		} else {
			i = crypto.CRC32(e.Session.SID()) % queueNum
		}
	case *ExecutorEvent:
		i = int(e.Event.UniqueId() % int64(queueNum))
	case *ExecutorRemote:
		i = rand.Intn(queueNum)
	}
	return i
}
