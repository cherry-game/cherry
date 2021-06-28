package cherryHandler

import (
	crypto "github.com/cherry-game/cherry/extend/crypto"
	reflect "github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
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
			handler.SetName(reflect.GetStructName(handler))
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

func DefaultQueueHash(executor IExecutor, queueNum int) int {
	var i = 0

	switch e := executor.(type) {
	case *MessageExecutor:
		if e.Agent.Session.UID() > 0 {
			i = crypto.CRC32(string(e.Agent.Session.UID())) % queueNum
		} else {
			i = crypto.CRC32(string(e.Agent.Session.SID())) % queueNum
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
