package cherryHandler

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"reflect"
)

type (
	ExecutorLocal struct {
		facade.IApplication
		Session      *cherrySession.Session
		Msg          *cherryMessage.Message
		HandlerFn    *facade.HandlerFn
		Ctx          context.Context
		AfterFilters []FilterFn
	}
)

func (m *ExecutorLocal) Invoke() {
	argsLen := len(m.HandlerFn.InArgs)
	if argsLen < 2 || argsLen > 3 {
		cherryLogger.Warnf("[Route = %v] method in args error.", m.Msg.Route)
		cherryLogger.Warnf("func(session,request) or func(ctx,session,request)")
		return
	}

	var params []reflect.Value

	if m.HandlerFn.IsRaw {
		if argsLen == 2 {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(m.Session)
			params[1] = reflect.ValueOf(m.Msg)
		} else {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(m.Ctx)
			params[1] = reflect.ValueOf(m.Session)
			params[2] = reflect.ValueOf(m.Msg)
		}
	} else {
		val, err := m.unmarshalData(argsLen - 1)
		if err != nil {
			cherryLogger.Warn(err)
			return
		}

		if argsLen == 2 {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(m.Session)
			params[1] = reflect.ValueOf(val)
		} else {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(m.Ctx)
			params[1] = reflect.ValueOf(m.Session)
			params[2] = reflect.ValueOf(val)
		}
	}

	ret := m.HandlerFn.Value.Call(params)

	if m.Msg.Type == cherryMessage.Request && len(ret) == 2 {
		if ret[0].IsNil() == false {
			m.Session.ResponseMID(m.Msg.ID, ret[0].Interface())
		} else if e := ret[1].Interface(); e != nil {
			if code, ok := e.(int32); ok {

				statusCode := cherryCode.GetCodeResult(code)
				m.Session.ResponseMID(m.Msg.ID, statusCode, true)
			} else {
				m.Session.Warn(e)
			}
		}
	}

	for _, filter := range m.AfterFilters {
		if !filter(m.Ctx, m.Session, m.Msg) {
			break
		}
	}
}

func (m *ExecutorLocal) unmarshalData(index int) (interface{}, error) {
	in2 := m.HandlerFn.InArgs[index]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()

	err := m.Unmarshal(m.Msg.Data, val)
	if err != nil {
		return nil, err
	}

	return val, err
}

func (m *ExecutorLocal) String() string {
	return m.Msg.Route
}
