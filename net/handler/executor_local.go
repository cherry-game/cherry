package cherryHandler

import (
	"context"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cmsg "github.com/cherry-game/cherry/net/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	csession "github.com/cherry-game/cherry/net/session"
	"go.uber.org/zap/zapcore"
	"reflect"
	"runtime/debug"
)

type (
	ExecutorLocal struct {
		cfacade.IApplication
		groupIndex    int
		session       *csession.Session
		msg           *cmsg.Message
		handlerFn     *cfacade.HandlerFn
		ctx           context.Context
		beforeFilters []FilterFn
		afterFilters  []FilterFn
	}
)

func (p *ExecutorLocal) Index() int {
	return p.groupIndex
}

func (p *ExecutorLocal) SetIndex(index int) {
	p.groupIndex = index
}

func (p *ExecutorLocal) Invoke() {
	defer func() {
		if rev := recover(); rev != nil {
			clog.Warnf("recover in Local. %s", string(debug.Stack()))
			clog.Warnf("msg = [%+v]", p.msg)
		}
	}()

	for _, filter := range p.beforeFilters {
		if filter(p.ctx, p.session, p.msg) == false {
			return
		}
	}

	argsLen := len(p.handlerFn.InArgs)
	if argsLen < 2 || argsLen > 3 {
		clog.Warnf("[Route = %v] method in args error.", p.msg.Route)
		clog.Warnf("func(session,request) or func(ctx,session,request)")
		return
	}

	var params []reflect.Value
	var ret []reflect.Value

	if p.handlerFn.IsRaw {
		params = make([]reflect.Value, argsLen)
		params[0] = reflect.ValueOf(p.ctx)
		params[1] = reflect.ValueOf(p.session)
		params[2] = reflect.ValueOf(p.msg)

		ret = p.handlerFn.Value.Call(params)
		if clog.PrintLevel(zapcore.DebugLevel) {
			p.session.Debugf("[local] [groupIndex = %d, route = %s, mid = %d, req = %+v, rsp = %+v]",
				p.groupIndex,
				p.msg.Route,
				p.msg.ID,
				p.msg.Data,
				printRet(ret),
			)
		}
	} else {
		val, err := p.unmarshalData(argsLen - 1)
		if err != nil {
			p.session.Errorf("err = %v, msg = %v", err, p.msg)
			return
		}

		if argsLen == 2 {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(p.session)
			params[1] = reflect.ValueOf(val)
		} else {
			params = make([]reflect.Value, argsLen)
			params[0] = reflect.ValueOf(p.ctx)
			params[1] = reflect.ValueOf(p.session)
			params[2] = reflect.ValueOf(val)
		}

		ret = p.handlerFn.Value.Call(params)
		if clog.PrintLevel(zapcore.DebugLevel) {
			p.session.Debugf("[local] [groupIndex = %d, route = %s, mid = %d, req = %+v, rsp = %+v]",
				p.groupIndex,
				p.msg.Route,
				p.msg.ID,
				val,
				printRet(ret),
			)
		}
	}

	if p.msg.Type == cmsg.Request {
		retLen := len(ret)

		if retLen == 2 {
			if ret[0].IsNil() {
				if face := ret[1].Interface(); face != nil {
					if code, ok := face.(int32); ok {
						rsp := &cproto.Response{
							Code: code,
						}

						p.session.ResponseMID(p.msg.ID, rsp, true)
					} else {
						p.session.Warn(face)
					}
				} else {
					p.session.Warnf("ret value type error. [type = %+v]", ret)
				}
			} else {
				p.session.ResponseMID(p.msg.ID, ret[0].Interface())
			}
		}
	}

	for _, filter := range p.afterFilters {
		if !filter(p.ctx, p.session, p.msg) {
			break
		}
	}
}

func (p *ExecutorLocal) unmarshalData(index int) (interface{}, error) {
	in2 := p.handlerFn.InArgs[index]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()

	err := p.Unmarshal(p.msg.Data, val)
	if err != nil {
		return nil, err
	}

	return val, err
}

func (p *ExecutorLocal) String() string {
	return p.msg.Route
}
