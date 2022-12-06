package cherryHandler

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	ccrypto "github.com/cherry-game/cherry/extend/crypto"
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
		Executor
		cfacade.IApplication
		session       *csession.Session
		msg           *cmsg.Message
		handlerFn     *cfacade.MethodInfo
		ctx           context.Context
		beforeFilters []FilterFn
		afterFilters  []FilterFn
	}
)

func (p *ExecutorLocal) Session() *csession.Session {
	return p.session
}

func (p *ExecutorLocal) Message() *cmsg.Message {
	return p.msg
}

func (p *ExecutorLocal) Context() context.Context {
	return p.ctx
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
	var rets []reflect.Value

	if p.handlerFn.IsRaw {
		params = make([]reflect.Value, argsLen)
		params[0] = reflect.ValueOf(p.ctx)
		params[1] = reflect.ValueOf(p.session)
		params[2] = reflect.ValueOf(p.msg)

		rets = p.handlerFn.Value.Call(params)
		if clog.PrintLevel(zapcore.DebugLevel) {
			p.session.Debugf("[local] [groupIndex = %d, route = %s, mid = %d, req = %+v, rsp = %+v]",
				p.groupIndex,
				p.msg.Route,
				p.msg.ID,
				p.msg.Data,
				rets,
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

		rets = p.handlerFn.Value.Call(params)
		if clog.PrintLevel(zapcore.DebugLevel) {
			p.session.Debugf("[local] [groupIndex = %d, route = %s, mid = %d, req = %+v, rsp = %+v]",
				p.groupIndex,
				p.msg.Route,
				p.msg.ID,
				val,
				rets,
			)
		}
	}

	if p.msg.Type == cmsg.Request {
		retLen := len(rets)

		if retLen == 1 {
			p.responseCode(rets[0])
		} else if retLen == 2 {
			if rets[0].IsNil() {
				p.responseCode(rets[1])
			} else {
				p.session.ResponseMID(p.msg.ID, rets[0].Interface())
			}
		} else {
			p.session.Warnf("[local] response type error. [route = %s, ret = %+v]", p.msg.Route, rets)
		}
	}

	p.executeAfterFilters()
}

func (p *ExecutorLocal) responseCode(ret reflect.Value) {
	if v := ret.Interface(); v != nil {
		if c, ok := v.(int32); ok {
			rsp := &cproto.Response{
				Code: c,
			}
			p.session.ResponseMID(p.msg.ID, rsp, cherryCode.IsFail(c))
		} else {
			p.session.Warn(v)
		}
	} else {
		p.session.Warnf("[local] response type error. [route = %s, ret = %+v]", p.msg.Route, ret)
	}
}

func (p *ExecutorLocal) executeAfterFilters() {
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

func (p *ExecutorLocal) QueueHash(queueNum int) int {
	if p.session.UID() > 0 {
		return int(p.session.UID() % int64(queueNum))
	}

	return ccrypto.CRC32(p.session.SID()) % queueNum
}
