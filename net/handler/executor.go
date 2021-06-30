package cherryHandler

import (
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"reflect"
)

type (
	IExecutor interface {
		Invoke()
		String() string
	}

	MessageExecutor struct {
		App           facade.IApplication
		Session       *cherrySession.Session
		Msg           *cherryMessage.Message
		HandlerFn     *facade.HandlerFn
		BeforeFilters []FilterFn
		AfterFilters  []FilterFn
	}

	EventExecutor struct {
		Event   facade.IEvent
		EventFn []facade.EventFn
	}

	UserRPCExecutor struct {
		HandlerFn *facade.HandlerFn
	}

	SysRPCExecutor struct {
		HandlerFn *facade.HandlerFn
	}
)

func (m *MessageExecutor) Invoke() {
	if m.Msg.Data == nil {
		cherryLogger.Warnf("[Route = %v] message data is nil", m.Msg.Route)
		return
	}

	for _, filter := range m.BeforeFilters {
		if !filter(m) {
			break
		}
	}

	argsLen := len(m.HandlerFn.InArgs)
	if argsLen < 2 || argsLen > 3 {
		cherryLogger.Warnf("[Route = %v] method in args error. func(session,message,data) or func(session,message)", m.Msg.Route)
		return
	}

	val, err := m.unmarshalData(argsLen - 1)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	var params []reflect.Value

	if argsLen == 2 {
		params = make([]reflect.Value, argsLen)
		params[0] = reflect.ValueOf(m.Session)
		params[1] = reflect.ValueOf(m.Msg)
	}

	if argsLen == 3 {
		params = make([]reflect.Value, argsLen)
		params[0] = reflect.ValueOf(m.Session)
		params[1] = reflect.ValueOf(m.Msg)
		params[2] = reflect.ValueOf(val)
	}

	result := m.HandlerFn.Value.Call(params)
	if len(result) > 0 {
		if err := result[0].Interface(); err != nil {
			cherryLogger.Warn(err)
			return
		}
	}

	for _, filter := range m.AfterFilters {
		if !filter(m) {
			break
		}
	}
}

func (m *MessageExecutor) unmarshalData(index int) (interface{}, error) {
	in2 := m.HandlerFn.InArgs[index]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()

	err := m.App.Unmarshal(m.Msg.Data, val)
	if err != nil {
		return nil, err
	}

	return val, err
}

func (m *MessageExecutor) call(params []reflect.Value) {
	result := m.HandlerFn.Value.Call(params)
	if len(result) > 0 {
		if err := result[0].Interface(); err != nil {
			cherryLogger.Warn(err)
		}
	}
}

func (m *MessageExecutor) String() string {
	return m.Msg.Route
}

func (e *EventExecutor) Invoke() {
	for _, fn := range e.EventFn {
		fn(e.Event)
	}
}

func (e *EventExecutor) String() string {
	return e.Event.EventName()
}

func (u *UserRPCExecutor) Invoke() {

}

func (u *UserRPCExecutor) String() string {
	return ""
}

func (s *SysRPCExecutor) Invoke() {

}

func (s *SysRPCExecutor) String() string {
	return ""
}
