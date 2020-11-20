package cherryInterfaces

import "reflect"

type IHandler interface {
	IAppContext

	Name() string

	SetName(name string)

	Init()

	WorkerSize() uint

	SetWorkerSize(size uint)

	QueueSize() uint

	SetQueueSize(size uint)

	GetEvents() map[string][]EventFunc

	GetEvent(name string) ([]EventFunc, bool)

	GetLocals() map[string]*InvokeFunc

	GetLocal(funcName string) (*InvokeFunc, bool)

	GetRemotes() map[string]*InvokeFunc

	GetRemote(funcName string) (*InvokeFunc, bool)

	GetWorkerExecuteFunc() WorkerExecuteFunc

	WorkerHashFunc(message interface{}) uint

	Stop()
}

type InvokeFunc struct {
	Type    reflect.Type
	Value   reflect.Value
	InArgs  []reflect.Type
	OutArgs []reflect.Type
}

type EventFunc func(e IEvent)

type WorkerExecuteFunc func(handler IHandler, index int, msgChan chan interface{})
