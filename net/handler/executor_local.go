package cherryHandler

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	facade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"reflect"
	"runtime/debug"
)

type (
	ExecutorLocal struct {
		facade.IApplication
		Session       *cherrySession.Session
		Msg           *cherryMessage.Message
		HandlerFn     *facade.HandlerFn
		Ctx           context.Context
		BeforeFilters []FilterFn
		AfterFilters  []FilterFn
	}
)

func (p *ExecutorLocal) Invoke() {
	defer func() {
		if rev := recover(); rev != nil {
			cherryLogger.Warnf("recover in Local. %s", string(debug.Stack()))
			cherryLogger.Warnf("msg = [%+v]", p.Msg)
		}
	}()

	for _, filter := range p.BeforeFilters {
		if filter(p.Ctx, p.Session, p.Msg) == false {
			return
		}
	}

	argsLen := len(p.HandlerFn.InArgs)
	if argsLen < 2 || argsLen > 3 {
		cherryLogger.Warnf("[Route = %v] method in args error.", p.Msg.Route)
		cherryLogger.Warnf("func(session,request) or func(ctx,session,request)")
		return
	}

	var params []reflect.Value

	if p.HandlerFn.IsRaw {
		if argsLen == 2 {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(p.Session)
			params[1] = reflect.ValueOf(p.Msg)
		} else {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(p.Ctx)
			params[1] = reflect.ValueOf(p.Session)
			params[2] = reflect.ValueOf(p.Msg)
		}
	} else {
		val, err := p.unmarshalData(argsLen - 1)
		if err != nil {
			cherryLogger.Warnf("err = %v, msg = %v", err, p.Msg)
			return
		}

		if argsLen == 2 {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(p.Session)
			params[1] = reflect.ValueOf(val)
		} else {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(p.Ctx)
			params[1] = reflect.ValueOf(p.Session)
			params[2] = reflect.ValueOf(val)
		}
	}

	ret := p.HandlerFn.Value.Call(params)

	if p.Msg.Type == cherryMessage.Request && len(ret) == 2 {
		if ret[0].IsNil() == false {
			p.Session.ResponseMID(p.Msg.ID, ret[0].Interface())
		} else if e := ret[1].Interface(); e != nil {
			if code, ok := e.(int32); ok {

				statusCode := cherryCode.GetCodeResult(code)
				p.Session.ResponseMID(p.Msg.ID, statusCode, true)
			} else {
				p.Session.Warn(e)
			}
		}
	}

	for _, filter := range p.AfterFilters {
		if !filter(p.Ctx, p.Session, p.Msg) {
			break
		}
	}
}

func (p *ExecutorLocal) unmarshalData(index int) (interface{}, error) {
	in2 := p.HandlerFn.InArgs[index]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()

	err := p.Unmarshal(p.Msg.Data, val)
	if err != nil {
		return nil, err
	}

	return val, err
}

func (p *ExecutorLocal) String() string {
	return p.Msg.Route
}
