package cherryInterfaces

import "reflect"

type IHandler interface {
	IAppContext

	Name() string

	SetName(name string)

	PreInit()

	Init()

	AfterInit()

	Events() map[string][]EventFn

	Event(name string) ([]EventFn, bool)

	LocalHandlers() map[string]*InvokeFn

	LocalHandler(funcName string) (*InvokeFn, bool)

	RemoteHandlers() map[string]*InvokeFn

	RemoteHandler(funcName string) (*InvokeFn, bool)

	PutMessage(message interface{})

	Stop()
}

type InvokeFn struct {
	Type    reflect.Type
	Value   reflect.Value
	InArgs  []reflect.Type
	OutArgs []reflect.Type
}

type EventFn func(e IEvent)
