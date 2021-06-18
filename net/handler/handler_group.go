package cherryHandler

import (
	cherryReflect "github.com/cherry-game/cherry/extend/reflect"
	facade "github.com/cherry-game/cherry/facade"
)

type (
	HandlerGroup struct {
		handlers  map[string]facade.IHandler
		queueNum  int
		queueCap  int
		queueMaps map[int]*Queue
		msgHash   func(msg *MessageExecutor, queueNum int) int
		eventHash func(msg *EventExecutor, queueNum int) int
	}

	Queue struct {
		index    int
		dataChan chan IExecutor
	}
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

func (h *HandlerGroup) SetEventHash(fn func(msg *EventExecutor, queueNum int) int) {
	h.eventHash = fn
}

func (h *HandlerGroup) SetMsgHash(fn func(msg *MessageExecutor, queueNum int) int) {
	h.msgHash = fn
}

func (h *HandlerGroup) inQueue(index int, executor IExecutor) {
	if index > h.queueNum {
		index = 0
	}

	q := h.queueMaps[index]
	q.dataChan <- executor
}

//func (h *HandlerGroup) SetQueueHash(hashFn QueueHashFn) {
//	if hashFn == nil {
//		logger.Warn("hashFn is nil")
//		return
//	}
//
//	h.queueHashFn = hashFn
//}
//
//func (h *HandlerGroup) SetQueueRandHash() {
//	h.SetQueueHash(func(queueNum int, msg interface{}) int {
//		return rand.Int() % queueNum
//	})
//}
//
//func (h *HandlerGroup) SetQueueCRC32Hash() {
//	h.SetQueueHash(func(queueNum int, msg interface{}) int {
//		var hashValue string
//		switch m := msg.(type) {
//		case facade.IEvent:
//			{
//				hashValue = m.UniqueId()
//			}
//		case *UnhandledMessage:
//			{
//				if m.Session != nil {
//					hashValue = m.Session.UID()
//				}
//			}
//		}
//
//		// default index
//		if hashValue == "" {
//			return 0
//		}
//
//		return crypto.CRC32(hashValue) % queueNum
//	})
//}

//func (h *HandlerGroup) SetQueueExecutor(executor QueueExecuteFn) {
//	h.queueExecuteFn = executor
//}

//func DefaultQueueExecutor(component *Component, group *HandlerGroup, queue *Queue) {
//	defer func() {
//		if r := recover(); r != nil {
//			logger.Warnf("recover in ProcessMessage(). %s", r)
//		}
//	}()
//
//	for {
//		select {
//		case message := <-queue.dataChan:
//			{
//				switch msg := message.(type) {
//				case *UnhandledMessage:
//					{
//						for _, filter := range component.GetBeforeFilter() {
//							if !filter(msg) {
//								break
//							}
//						}
//
//						for _, handler := range group.handlers {
//							if method, found := handler.LocalHandler(msg.Route.Method()); found {
//
//								if profile.Debug() {
//									logger.Debugf("[%s-chan-%d] receive message. route = %s",
//										handler.Name(), queue.Index, msg.Route.String())
//								}
//
//								params := make([]reflect.Value, 2)
//								params[0] = reflect.ValueOf(msg.Session)
//								params[1] = reflect.ValueOf(msg.Msg)
//								method.Value.Call(params)
//							}
//						}
//
//						for _, filter := range component.GetAfterFilter() {
//							if !filter(msg) {
//								break
//							}
//						}
//					}
//				case facade.IEvent:
//					{
//						for _, handler := range group.handlers {
//							calls, found := handler.Event(msg.EventName())
//							if found == false {
//								break
//							}
//
//							if profile.Debug() {
//								logger.Debugf("[%s-chan-%d] receive event. msg type = %v",
//									handler.Name(), queue.Index, reflect.TypeOf(message))
//							}
//
//							for _, call := range calls {
//								call(msg)
//							}
//						}
//					}
//				default:
//					{
//						logger.Warnf("message not process. value = %v", msg)
//					}
//				}
//			}
//			//case timer
//		}
//	}
//}
