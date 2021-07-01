package cherryHandler

import (
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"reflect"
)

type (
	MessageExecutor struct {
		App           facade.IApplication
		Session       *cherrySession.Session
		Msg           *cherryMessage.Message
		HandlerFn     *facade.HandlerFn
		BeforeFilters []FilterFn
		AfterFilters  []FilterFn
	}
)

func (m *MessageExecutor) Invoke() {
	if m.Msg.Data == nil {
		cherryLogger.Warnf("[Route = %v] message data is nil", m.Msg.Route)
		return
	}

	for _, filter := range m.BeforeFilters {
		if filter(m.Session, m.Msg) == false {
			return
		}
	}

	argsLen := len(m.HandlerFn.InArgs)
	if argsLen < 2 || argsLen > 3 {
		cherryLogger.Warnf("[Route = %v] method in args error. func(session,message,data) or func(session,message)", m.Msg.Route)
		return
	}

	var params []reflect.Value

	if argsLen == 2 {
		params = make([]reflect.Value, argsLen)
		params[0] = reflect.ValueOf(m.Session)
		params[1] = reflect.ValueOf(m.Msg)
	}

	if argsLen == 3 {
		val, err := m.unmarshalData(argsLen - 1)
		if err != nil {
			cherryLogger.Warn(err)
			return
		}

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
		if !filter(m.Session, m.Msg) {
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
