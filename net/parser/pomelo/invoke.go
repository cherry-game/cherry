package pomelo

import (
	ccode "github.com/cherry-game/cherry/code"
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/gogo/protobuf/proto"
	"reflect"
)

// local  func format : funcName(session *Session,req *Request)
// remote func format : funcName()
// remote func format : funcName(req *Request)
// remote func format : funcName(req *Request) (*Rsp,int32)

var (
	localFuncTypes []reflect.Type
	localFuncLen   = 2
)

func init() {
	localFuncTypes = make([]reflect.Type, localFuncLen)
	localFuncTypes[0] = reflect.TypeOf(&cproto.Session{})
	localFuncTypes[1] = reflect.TypeOf(&pomeloMessage.Message{})
}

func LocalInvokeFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if m.IsCluster {
		localCluster(app, fi, m)
	} else {
		local(app, fi, m)
	}
}

func local(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if len(fi.InArgs) != localFuncLen {
		clog.Errorf("[local] args error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
		return
	}

	msgData, ok := m.Args[1].([]byte)
	if !ok {
		clog.Errorf("[local] Message arg error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
		return
	}

	var arg1Value interface{}
	arg1Value = reflect.New(fi.InArgs[1].Elem()).Interface()
	err := app.Serializer().Unmarshal(msgData, arg1Value)
	if err != nil {
		clog.Errorf("[local] args index of 2 unmarshal error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
		return
	}

	values := make([]reflect.Value, localFuncLen)
	values[0] = reflect.ValueOf(m.Args[0])
	values[1] = reflect.ValueOf(arg1Value)

	fi.Value.Call(values)
}

func localCluster(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if m.Args == nil || len(m.Args) != 2 {
		return
	}

	session, ok := m.Args[0].(*cproto.Session)
	if !ok {
		clog.Errorf("[localCluster] Decode session error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
		return
	}

	arg1Bytes, ok := m.Args[1].([]byte)
	if !ok {
		clog.Errorf("[localCluster] argBytes type error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
		return
	}

	var arg1Value interface{}
	arg1Value = reflect.New(fi.InArgs[1].Elem()).Interface()
	err := app.Serializer().Unmarshal(arg1Bytes, arg1Value)
	if err != nil {
		clog.Errorf("[localCluster] Unmarshal error.[source = %s, target = %s -> %s, funcType = %v, err = %+v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
			err,
		)
		return
	}

	values := make([]reflect.Value, 2)
	values[0] = reflect.ValueOf(session)
	values[1] = reflect.ValueOf(arg1Value)
	fi.Value.Call(values)
}

func RemoteInvokeFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if m.IsCluster {
		remoteCluster(app, fi, m)
	} else {
		remote(app, fi, m)
	}
}

func remote(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	values := make([]reflect.Value, fi.InArgsLen)
	for i, arg := range m.Args {
		values[i] = reflect.ValueOf(arg)
	}

	cutils.Try(func() {
		if m.ChanResult == nil {
			fi.Value.Call(values)
		} else {
			rets := fi.Value.Call(values)
			rsp := ret2Response(app.Serializer(), rets)
			m.ChanResult <- &rsp
		}
	}, func(errString string) {
		if m.ChanResult != nil {
			m.ChanResult <- nil
		}

		clog.Errorf("[remote] invoke error.[source = %s, target = %s -> %s, funcType = %v, err = %+v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
			errString,
		)
	})
}

func remoteCluster(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if fi.InArgsLen == 0 {
		values := make([]reflect.Value, fi.InArgsLen)
		cutils.Try(func() {
			rets := fi.Value.Call(values)
			rsp := ret2Response(app.Serializer(), rets)
			returnResponse(m.ClusterReply, &rsp)
		}, func(errString string) {
			clog.Warn(errString)
			returnResponse(m.ClusterReply, &cproto.Response{
				Code: ccode.RPCRemoteExecuteError,
			})
		})
		return
	}

	argBytes, ok := m.Args[0].([]byte)
	if !ok {
		clog.Errorf("[remoteCluster] argBytes type error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
		return
	}

	var argValue interface{}
	argValue = reflect.New(fi.InArgs[0].Elem()).Interface()
	err := app.Serializer().Unmarshal(argBytes, argValue)
	if err != nil {
		clog.Errorf("[remoteCluster] Decode argBytes error.[source = %s, target = %s -> %s, funcType = %v, err = %+v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
			err,
		)
		return
	}

	values := make([]reflect.Value, len(m.Args))
	values[0] = reflect.ValueOf(argValue)

	cutils.Try(func() {
		rets := fi.Value.Call(values)
		rsp := ret2Response(app.Serializer(), rets)
		returnResponse(m.ClusterReply, &rsp)
	}, func(errString string) {
		returnResponse(m.ClusterReply, &cproto.Response{
			Code: ccode.RPCRemoteExecuteError,
		})

		clog.Errorf("[remoteCluster] invoke error.[source = %s, target = %s -> %s, funcType = %v, err = %+v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
			errString,
		)
	})
}

func ret2Response(serializer cfacade.ISerializer, rets []reflect.Value) cproto.Response {
	rsp := cproto.Response{
		Code: ccode.OK,
	}

	retsLen := len(rets)
	if retsLen == 1 {
		if val := rets[0].Interface(); val != nil {
			if c, ok := val.(int32); ok {
				rsp.Code = c
			}
		}
	} else if retsLen == 2 {
		if !rets[0].IsNil() {
			data, err := serializer.Marshal(rets[0].Interface())
			if err != nil {
				rsp.Code = ccode.RPCRemoteExecuteError
				clog.Warn(err)
			} else {
				rsp.Data = data
			}
		}

		if val := rets[1].Interface(); val != nil {
			if c, ok := val.(int32); ok {
				rsp.Code = c
			}
		}
	}

	return rsp
}

func returnResponse(reply cfacade.IRespond, rsp *cproto.Response) {
	if reply != nil {
		rspData, _ := proto.Marshal(rsp)
		err := reply.Respond(rspData)
		if err != nil {
			clog.Warn(err)
		}
	}
}
