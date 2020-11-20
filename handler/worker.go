package cherryHandler

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"reflect"
)

type Worker struct {
	cherryInterfaces.IHandler
	handlerComponent *HandlerComponent
	messageChan      []chan interface{} // message chan slice
}

func NewWorker(component *HandlerComponent, handler *Handler) *Worker {
	w := &Worker{
		handlerComponent: component,
		IHandler:         handler,
	}

	w.messageChan = make([]chan interface{}, handler.WorkerSize())

	for i := 0; i < int(handler.WorkerSize()); i++ {
		w.messageChan[i] = make(chan interface{}, handler.QueueSize())
	}

	return w
}

func (h *Worker) Start() {
	f := h.GetWorkerExecuteFunc()

	if f == nil {
		f = h.DefaultExecuteWorker
	}

	for i := 0; i < int(h.WorkerSize()); i++ {
		go f(h.IHandler, i, h.messageChan[i])
	}
}

func (h *Worker) Stop() {
	//处理完chan的消息，再退出for

	h.IHandler.Stop()
}

func (h *Worker) PutMessage(message interface{}) {
	var index uint
	if h.WorkerSize() > 1 {
		index = h.WorkerHashFunc(message)
	}

	if index > h.WorkerSize() {
		index = 0
	}

	h.messageChan[index] <- message
}

func (h *Worker) DefaultExecuteWorker(handler cherryInterfaces.IHandler, chanIndex int, msgChan chan interface{}) {
	for {
		select {
		case msg, found := <-msgChan:
			{
				if !found && !handler.App().Running() {
					return
				}

				h.ProcessMessage(handler, chanIndex, msg)
			}
		}
		//case timer
	}
}

func (h *Worker) ProcessMessage(handler cherryInterfaces.IHandler, chanIndex int, message interface{}) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("recover in ProcessMessage(). %s", r)
		}
	}()

	switch msg := message.(type) {
	case UnhandledMessage:
		{
			for _, filter := range h.handlerComponent.GetBeforeFilter() {
				if !filter(msg) {
					break
				}
			}

			if method, found := handler.GetLocal(msg.Route.Method()); found {
				params := make([]reflect.Value, 2)
				params[0] = reflect.ValueOf(msg.Session)
				params[1] = reflect.ValueOf(msg.Msg)
				method.Value.Call(params)
			}

			if cherryProfile.Debug() {
				cherryLogger.Debugf("[%s-chan-%d] receive message. route = %s",
					handler.Name(), chanIndex, msg.Route.String())
			}

			for _, filter := range h.handlerComponent.GetAfterFilter() {
				if !filter(msg) {
					break
				}
			}
		}
	case cherryInterfaces.IEvent:
		{
			if cherryProfile.Debug() {
				cherryLogger.Debugf("[%s-chan-%d] receive event. msg type = %v",
					handler.Name(), chanIndex, reflect.TypeOf(message))
			}

			calls, found := handler.GetEvent(msg.EventName())
			if found == false {
				return
			}

			for _, call := range calls {
				call(msg)
			}
		}
	default:
		{
			cherryLogger.Warnf("message not process. value = %s", msg)
		}
	}
}
