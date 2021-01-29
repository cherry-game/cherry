package cherryHandler

import (
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"hash/crc32"
	"math/rand"
	"reflect"
)

type (
	WorkerExecuteFn func(handler cherryInterfaces.IHandler, index int, msgChan chan interface{})
	WorkerHashFn    func(msg interface{}) uint

	Worker struct {
		queueSize       uint               // chan size
		workerSize      uint               // worker size
		messageChan     []chan interface{} // message chan slice
		workerHashFn    WorkerHashFn       // goroutine hash function
		workerExecuteFn WorkerExecuteFn    // worker execute function
	}
)

func (w *Worker) SetQueueSize(size uint) {
	w.queueSize = size
}

func (w *Worker) WorkerHash(workerSize uint, hashFn WorkerHashFn) {
	if workerSize > 1 && hashFn == nil {
		cherryLogger.Warn("WorkerHashFn is nil")
		return
	}

	w.workerSize = workerSize
	w.workerHashFn = hashFn
}

func (w *Worker) WorkerRandHash(workerSize uint) {
	w.WorkerHash(workerSize, func(_ interface{}) uint {
		return uint(rand.Uint32()) % workerSize
	})
}

func (w *Worker) WorkerCRC32Hash(workerSize uint) {
	w.WorkerHash(workerSize, func(msg interface{}) uint {
		var hashValue string
		switch m := msg.(type) {
		case cherryInterfaces.IEvent:
			{
				hashValue = m.UniqueId()
			}
		case UnhandledMessage:
			{
				if m.Session != nil {
					hashValue = string(m.Session.UID())
				}
			}
		}

		if hashValue == "" {
			return 0
		}

		return uint(crc32.ChecksumIEEE([]byte(hashValue))) % workerSize
	})
}

func (w *Worker) DefaultExecuteWorker(handler cherryInterfaces.IHandler, chanIndex int, msgChan chan interface{}) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("recover in ProcessMessage(). %s", r)
		}
	}()

	component := handler.App().Find(cherryConst.HandlerComponent).(*HandlerComponent)
	if component == nil {
		cherryLogger.Warn("not find HandlerComponent.")
		return
	}

	for {
		select {
		case message, found := <-msgChan:
			{
				if !found && !handler.App().Running() {
					return
				}

				switch msg := message.(type) {
				case *UnhandledMessage:
					{
						for _, filter := range component.GetBeforeFilter() {
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

						for _, filter := range component.GetAfterFilter() {
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
						cherryLogger.Warnf("message not process. value = %v", msg)
					}
				}
			}
		}
		//case timer
	}
}
