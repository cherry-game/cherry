package cherryHandler

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/crypto"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"math/rand"
	"reflect"
)

type (
	WorkerExecuteFn func(handler cherryFacade.IHandler, worker *Worker) // worker执行函数
	WorkerHashFn    func(msg interface{}) int                           // worker hash函数(根据规则找到具体的worker)

	// WorkerGroup
	WorkerGroup struct {
		queueSize        int             // chan size
		workerSize       int             // worker size
		workerMap        map[int]*Worker // workerMap key:index,value:Worker
		workerHashFn     WorkerHashFn    // goroutine hash function
		workerExecutorFn WorkerExecuteFn // worker execute function
	}

	// Worker 单个worker对象
	Worker struct {
		Index       int              // 索引id
		MessageChan chan interface{} // 消息对列
	}
)

func (w *WorkerGroup) initWorkerGroup() {
	if w.queueSize < 1 {
		w.queueSize = 32767
	}

	if w.workerSize < 1 {
		w.workerSize = 1
	}

	if w.workerExecutorFn == nil {
		w.workerExecutorFn = DefaultWorkerExecutor
	}

	w.workerMap = make(map[int]*Worker)

	for i := 0; i < w.workerSize; i++ {
		worker := &Worker{
			Index:       i,
			MessageChan: make(chan interface{}, w.queueSize),
		}
		w.workerMap[i] = worker
	}
}

func (w *WorkerGroup) runWorker(handler cherryFacade.IHandler) {
	for i := 0; i < w.workerSize; i++ {
		worker := w.workerMap[i]
		// new goroutine for worker
		go w.workerExecutorFn(handler, worker)
	}
}

func (w *WorkerGroup) WorkerMap() map[int]*Worker {
	return w.workerMap
}

func (w *WorkerGroup) SetQueueSize(size int) {
	if size < 1 {
		size = 32767
	}

	w.queueSize = size
}

func (w *WorkerGroup) SetWorkerExecutor(workerExecuteFn WorkerExecuteFn) {
	w.workerExecutorFn = workerExecuteFn
}

func (w *WorkerGroup) SetWorkerHash(workerSize int, hashFn WorkerHashFn) {
	if workerSize <= 1 {
		workerSize = 1
		hashFn = nil
	}

	if workerSize > 1 && hashFn == nil {
		cherryLogger.Warn("WorkerHashFn is nil")
		return
	}

	w.workerSize = workerSize
	w.workerHashFn = hashFn
}

func (w *WorkerGroup) SetWorkerRandHash(workerSize int) {
	w.SetWorkerHash(workerSize, func(_ interface{}) int {
		return rand.Int() % workerSize
	})
}

func (w *WorkerGroup) SetWorkerCRC32Hash(workerSize int) {
	w.SetWorkerHash(workerSize, func(msg interface{}) int {
		var hashValue string
		switch m := msg.(type) {
		case cherryFacade.IEvent:
			{
				hashValue = m.UniqueId()
			}
		case *UnhandledMessage:
			{
				if m.Session != nil {
					hashValue = m.Session.UID()
				}
			}
		}

		// default index
		if hashValue == "" {
			return 0
		}

		return cherryCrypto.CRC32(hashValue) % workerSize
	})
}

// DefaultWorkerExecutor
func DefaultWorkerExecutor(handler cherryFacade.IHandler, worker *Worker) {
	defer func() {
		if r := recover(); r != nil {
			cherryLogger.Warnf("recover in ProcessMessage(). %s", r)
		}
	}()

	component := handler.App().Find(cherryConst.HandlerComponent).(*Component)
	if component == nil {
		cherryLogger.Warn("not find handlerComponent.")
		return
	}

	for {
		select {
		case message := <-worker.MessageChan:
			{
				switch msg := message.(type) {
				case *UnhandledMessage:
					{
						for _, filter := range component.GetBeforeFilter() {
							if !filter(msg) {
								break
							}
						}

						if cherryProfile.Debug() {
							cherryLogger.Debugf("[%s-chan-%d] receive message. route = %s",
								handler.Name(), worker.Index, msg.Route.String())
						}

						if method, found := handler.LocalHandler(msg.Route.Method()); found {
							params := make([]reflect.Value, 2)
							params[0] = reflect.ValueOf(msg.Session)
							params[1] = reflect.ValueOf(msg.Msg)
							method.Value.Call(params)
						}

						for _, filter := range component.GetAfterFilter() {
							if !filter(msg) {
								break
							}
						}
					}
				case cherryFacade.IEvent:
					{
						if cherryProfile.Debug() {
							cherryLogger.Debugf("[%s-chan-%d] receive event. msg type = %v",
								handler.Name(), worker.Index, reflect.TypeOf(message))
						}

						calls, found := handler.Event(msg.EventName())
						if found == false {
							break
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
			//case timer
		}
	}
}
