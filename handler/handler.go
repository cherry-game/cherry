package cherryHandler

import (
	"github.com/phantacix/cherry/interfaces"
	"github.com/phantacix/cherry/logger"
	"github.com/phantacix/cherry/profile"
	"github.com/phantacix/cherry/utils"
	"hash/crc32"
	"math/rand"
	"reflect"
)

type (
	Handler struct {
		cherryInterfaces.AppContext
		name              string                                  // unique Handler name
		eventFunc         map[string][]cherryInterfaces.EventFunc // event func
		localHandlers     map[string]*cherryInterfaces.InvokeFunc // local invoke Handler functions
		remoteHandlers    map[string]*cherryInterfaces.InvokeFunc // remote invoke Handler functions
		workerSize        uint                                    // worker size
		queueSize         uint                                    // size of each channel
		workerHashFunc    func(msg interface{}) uint              // goroutine hash function
		workerExecuteFunc cherryInterfaces.WorkerExecuteFunc      // worker execute function
	}
)

func (h *Handler) Name() string {
	return h.name
}

func (h *Handler) SetName(name string) {
	h.name = name
}

func (h *Handler) Init() {
	if h.workerSize < 1 {
		h.workerSize = 1
	}

	if h.queueSize < 1 {
		h.queueSize = 1024
	}

	//if h.workerExecuteFunc == nil {
	//	h.workerExecuteFunc = h.DefaultExecuteWorker
	//}
	//
	//component := h.App().Find(cherryConst.HandlerComponent).(*HandlerComponent)
	//
	//h.Worker = NewWorker(component, h)
	//h.Worker.Start()
}

func (h *Handler) WorkerSize() uint {
	return h.workerSize
}

func (h *Handler) SetWorkerSize(size uint) {
	h.workerSize = size
}

func (h *Handler) QueueSize() uint {
	return h.queueSize
}

func (h *Handler) SetQueueSize(size uint) {
	h.queueSize = size
}

func (h *Handler) GetEvents() map[string][]cherryInterfaces.EventFunc {
	return h.eventFunc
}

func (h *Handler) GetEvent(name string) ([]cherryInterfaces.EventFunc, bool) {
	events, found := h.eventFunc[name]
	return events, found
}

func (h *Handler) GetLocals() map[string]*cherryInterfaces.InvokeFunc {
	return h.localHandlers
}

func (h *Handler) GetLocal(funcName string) (*cherryInterfaces.InvokeFunc, bool) {
	invoke, found := h.localHandlers[funcName]
	return invoke, found
}

func (h *Handler) GetRemotes() map[string]*cherryInterfaces.InvokeFunc {
	return h.remoteHandlers
}

func (h *Handler) GetRemote(funcName string) (*cherryInterfaces.InvokeFunc, bool) {
	invoke, found := h.remoteHandlers[funcName]
	return invoke, found
}

func (h *Handler) GetWorkerExecuteFunc() cherryInterfaces.WorkerExecuteFunc {
	return h.workerExecuteFunc
}

func (h *Handler) WorkerHashFunc(message interface{}) uint {
	return h.workerHashFunc(message)
}

func (h *Handler) Stop() {

}

func (h *Handler) RegisterLocals(funcSlice ...interface{}) {
	for _, fn := range funcSlice {
		funcName := cherryUtils.Reflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.RegisterLocal(funcName, fn)
	}
}

func (h *Handler) RegisterLocal(name string, fn interface{}) {
	f, err := getInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	if h.localHandlers == nil {
		h.localHandlers = make(map[string]*cherryInterfaces.InvokeFunc)
	}

	h.localHandlers[name] = f

	cherryLogger.Debugf("[Handler = %s] register local func name = %s, numIn = %d, numOut =%d",
		h.name, name, len(f.InArgs), len(f.OutArgs))
}

func (h *Handler) RegisterRemotes(funcSlice ...interface{}) {
	for _, fn := range funcSlice {
		funcName := cherryUtils.Reflect.GetFuncName(fn)
		if funcName == "" {
			cherryLogger.Warnf("get function name fail. fn=%v", fn)
			continue
		}
		h.RegisterRemote(funcName, fn)
	}
}

func (h *Handler) RegisterRemote(name string, fn interface{}) {
	invokeFunc, err := getInvokeFunc(name, fn)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	if h.remoteHandlers == nil {
		h.remoteHandlers = make(map[string]*cherryInterfaces.InvokeFunc)
	}

	h.remoteHandlers[name] = invokeFunc

	cherryLogger.Debugf("[Handler = %s] register remote func name = %s, numIn = %d, numOut = %d",
		h.name, name, len(invokeFunc.InArgs), len(invokeFunc.OutArgs))
}

func (h *Handler) PostEvent(e cherryInterfaces.IEvent) {
	h.App().PostEvent(e)
}

//RegisterEvent
func (h *Handler) RegisterEvent(eventName string, fn cherryInterfaces.EventFunc) {
	if eventName == "" {
		cherryLogger.Warn("eventName is nil")
		return
	}

	if fn == nil {
		cherryLogger.Warn("event function is nil")
		return
	}

	if h.eventFunc == nil {
		h.eventFunc = make(map[string][]cherryInterfaces.EventFunc)
	}

	events := h.eventFunc[eventName]
	events = append(events, fn)

	h.eventFunc[eventName] = events

	if cherryProfile.Debug() {
		cherryLogger.Debugf("[Handler = %s] register event = %s.", h.name, eventName)
	}
}

func (h *Handler) WorkerHash(workerSize uint, hashFunc func(message interface{}) uint) {
	if workerSize < 1 && hashFunc == nil {
		return
	}

	h.workerSize = workerSize
	h.workerHashFunc = hashFunc
}

func (h *Handler) WorkerRandHash(workerSize uint) {
	h.WorkerHash(workerSize, func(_ interface{}) uint {
		return uint(rand.Uint32()) % h.workerSize
	})
}

func (h *Handler) WorkerCRC32Hash(workerSize uint) {
	h.WorkerHash(workerSize, func(msg interface{}) uint {
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
		return uint(crc32.ChecksumIEEE([]byte(hashValue))) % h.workerSize
	})
}

func (h *Handler) SetWorkerExecuteFunc(executeFunc cherryInterfaces.WorkerExecuteFunc) {
	if executeFunc != nil {
		h.workerExecuteFunc = executeFunc
	}
}

//func DefaultWorkerExecute(handler cherryInterfaces.IHandler, chanIndex int, msgChan chan interface{}) {
//	for {
//		select {
//		case msg, found := <-msgChan:
//			{
//				if !found && !handler.App().Running() {
//					break
//				}
//				ProcessMessage(handler, chanIndex, msg)
//			}
//		}
//		//case timer
//	}
//}

//func (h *Handler) PutMessage(message interface{}) {
//	//var index uint
//	//if h.workerSize > 1 {
//	//	index = h.workerHashFunc(message)
//	//}
//	//
//	//if index > h.workerSize {
//	//	index = 0
//	//}
//	//
//	//h.messageChan[index] <- message
//	h.messageChan <- message
//}

//getInvokeFunc reflect function convert to InvokeFunc
func getInvokeFunc(name string, fn interface{}) (*cherryInterfaces.InvokeFunc, error) {
	if name == "" {
		return nil, cherryUtils.Error("func name is nil")
	}

	if fn == nil {
		return nil, cherryUtils.ErrorFormat("func is nil. name = %s", name)
	}

	typ := reflect.TypeOf(fn)
	val := reflect.ValueOf(fn)

	if typ.Kind() != reflect.Func {
		return nil, cherryUtils.ErrorFormat("name = %s is not func type.", name)
	}

	var inArgs []reflect.Type
	for i := 0; i < typ.NumIn(); i++ {
		t := typ.In(i)
		inArgs = append(inArgs, t)
	}

	var outArgs []reflect.Type
	for i := 0; i < typ.NumOut(); i++ {
		t := typ.Out(i)
		outArgs = append(outArgs, t)
	}

	invoke := &cherryInterfaces.InvokeFunc{
		Type:    typ,
		Value:   val,
		InArgs:  inArgs,
		OutArgs: outArgs,
	}

	return invoke, nil
}
