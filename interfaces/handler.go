package cherryInterfaces

import "reflect"

type IHandler interface {
	IAppContext

	Name() string

	SetName(name string)

	Init()

	AfterInit()

	GetEvents() map[string][]EventFn

	GetEvent(name string) ([]EventFn, bool)

	GetLocals() map[string]*InvokeFn

	GetLocal(funcName string) (*InvokeFn, bool)

	GetRemotes() map[string]*InvokeFn

	GetRemote(funcName string) (*InvokeFn, bool)

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
