package cherryHandler

import (
	"fmt"
	ccode "github.com/cherry-game/cherry/code"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap/zapcore"
	"reflect"
	"runtime/debug"
)

type (
	ExecutorRemote struct {
		cfacade.IApplication
		groupIndex int
		handlerFn  *cfacade.HandlerFn
		route      string
		data       []byte
		natsMsg    *nats.Msg
	}
)

func NewExecutorRemote(route string, data []byte, natsMsg *nats.Msg) ExecutorRemote {
	return ExecutorRemote{
		route:   route,
		data:    data,
		natsMsg: natsMsg,
	}
}

func (p *ExecutorRemote) Index() int {
	return p.groupIndex
}

func (p *ExecutorRemote) SetIndex(index int) {
	p.groupIndex = index
}

func (p *ExecutorRemote) Invoke() {
	defer func() {
		if rev := recover(); rev != nil {
			clog.Warnf("[remote] recover in Remote. %s", string(debug.Stack()))
			clog.Warnf("[route = %s,data = %v]", p.route, p.data)
		}
	}()

	argsLen := len(p.handlerFn.InArgs)
	if argsLen < 0 || argsLen > 1 {
		clog.Warnf("[remote] method in args error. [route = %s]", p.route)
		clog.Warnf("func() or func(request)")
		return
	}

	var params []reflect.Value
	var ret []reflect.Value

	switch argsLen {
	case 0:
		{
			ret = p.handlerFn.Value.Call(params)
			if clog.LogLevel(zapcore.DebugLevel) {
				clog.Debugf("[remote] [route = %s, req = null, rsp = %+v",
					p.route,
					printRet(ret),
				)
			}
			break
		}
	case 1:
		{
			val, err := p.unmarshalData()
			if err != nil {
				clog.Warnf("[remote] unmarshal data error. [route = %s, error = %s]",
					p.route,
					err,
				)
				return
			}
			params = make([]reflect.Value, 1)
			params[0] = reflect.ValueOf(val)

			ret = p.handlerFn.Value.Call(params)
			if clog.LogLevel(zapcore.DebugLevel) {
				clog.Debugf("[remote] call result. [route = %s, req = %v, rsp = %+v]",
					p.route,
					params[0],
					printRet(ret),
				)
			}
			break
		}
	}

	if p.natsMsg.Reply == "" {
		return
	}

	rsp := &cproto.Response{
		Code: ccode.OK,
	}

	if len(ret) == 1 {
		if val := ret[0].Interface(); val != nil {
			if code, ok := val.(int32); ok {
				rsp.Code = code
			}
		}

		rspData, _ := p.Marshal(rsp)
		err := p.natsMsg.Respond(rspData)
		if err != nil {
			clog.Warn(err)
		}

	} else if len(ret) == 2 {
		if val := ret[1].Interface(); val != nil {
			if code, ok := val.(int32); ok {
				rsp.Code = code
			}
		}

		if ret[0].IsNil() == false {
			data, err := p.Marshal(ret[0].Interface())
			if err != nil {
				rsp.Code = ccode.RPCRemoteExecuteError
				clog.Warn(err)
			} else {
				rsp.Data = data
			}
		}

		rspData, _ := p.Marshal(rsp)
		err := p.natsMsg.Respond(rspData)
		if err != nil {
			clog.Warn(err)
		}
	}
}

func (p *ExecutorRemote) unmarshalData() (interface{}, error) {
	if len(p.handlerFn.InArgs) != 1 {
		return nil, cerr.Error("[remote] handler params len is error.")
	}

	in2 := p.handlerFn.InArgs[0]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()
	err := p.Unmarshal(p.data, val)
	if err != nil {
		return nil, err
	}

	return val, err
}

func (p *ExecutorRemote) String() string {
	return p.route
}

func printRet(t []reflect.Value) interface{} {
	switch len(t) {
	case 1:
		{
			return fmt.Sprintf("%v", t[0].Interface())
		}
	case 2:
		{
			return fmt.Sprintf("%v, %v", t[0].Interface(), t[0].Interface())
		}
	}

	return fmt.Sprint("")
}
