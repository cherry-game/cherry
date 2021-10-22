package cherryHandler

import (
	cherryCode "github.com/cherry-game/cherry/code"
	cherryError "github.com/cherry-game/cherry/error"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	"github.com/nats-io/nats.go"
	"reflect"
)

type (
	ExecutorRemote struct {
		facade.IApplication
		HandlerFn    *facade.HandlerFn
		RemotePacket *cherryProto.RemotePacket
		NatsMsg      *nats.Msg
	}
)

func (r *ExecutorRemote) Invoke() {
	argsLen := len(r.HandlerFn.InArgs)
	if argsLen < 0 || argsLen > 1 {
		cherryLogger.Warnf("[Route = %v] method in args error.", r.RemotePacket.Route)
		cherryLogger.Warnf("func() or func(request)")
		return
	}

	var ret []reflect.Value
	var params []reflect.Value

	switch argsLen {
	case 0:
		ret = r.HandlerFn.Value.Call(params)
		break
	case 1:
		val, err := r.unmarshalData()
		if err != nil {
			cherryLogger.Warnf("[Route = %s] unmarshal data error.error = %s", r.RemotePacket.Route, err)
			return
		}
		params = make([]reflect.Value, 1)
		params[0] = reflect.ValueOf(val)

		ret = r.HandlerFn.Value.Call(params)
		break
	}

	if r.NatsMsg.Reply == "" {
		return
	}

	rsp := &cherryProto.Response{
		Code: cherryCode.OK,
	}

	if len(ret) == 1 {
		if e := ret[0].Interface(); e != nil {
			if code, ok := e.(int32); ok {
				rsp.Code = code
			}
		}

		rspData, _ := r.Marshal(rsp)
		err := r.NatsMsg.Respond(rspData)
		if err != nil {
			cherryLogger.Warn(err)
		}

	} else if len(ret) == 2 {
		if e := ret[1].Interface(); e != nil {
			if code, ok := e.(int32); ok {
				rsp.Code = code
			}
		}

		if ret[0].IsNil() == false {

			data, err := r.Marshal(ret[0].Interface())
			if err != nil {
				rsp.Code = cherryCode.RPCRemoteExecuteError
				cherryLogger.Warn(err)
			} else {
				rsp.Data = data
			}
		}

		rspData, _ := r.Marshal(rsp)
		err := r.NatsMsg.Respond(rspData)
		if err != nil {
			cherryLogger.Warn(err)
		}
	}
}

func (r *ExecutorRemote) unmarshalData() (interface{}, error) {
	if len(r.HandlerFn.InArgs) != 1 {
		return nil, cherryError.Error("remote handler params len is error.")
	}

	in2 := r.HandlerFn.InArgs[0]

	var val interface{}
	val = reflect.New(in2.Elem()).Interface()
	err := r.Unmarshal(r.RemotePacket.Data, val)
	if err != nil {
		return nil, err
	}

	return val, err
}

func (r *ExecutorRemote) String() string {
	return r.RemotePacket.Route
}
