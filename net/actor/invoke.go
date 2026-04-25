package cherryActor

import (
	"reflect"

	"google.golang.org/protobuf/proto"

	ccode "github.com/cherry-game/cherry/code"
	cerror "github.com/cherry-game/cherry/error"
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cutils "github.com/cherry-game/cherry/extend/utils"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cnats "github.com/cherry-game/cherry/net/nats"
	cproto "github.com/cherry-game/cherry/net/proto"
)

func InvokeLocalFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if app == nil {
		clog.Errorf("[InvokeLocalFunc] app is nil. [message = %+v]", m)
		return
	}

	if err := EncodeLocalArgs(app, fi, m); err != nil {
		clog.Errorf("[InvokeLocalFunc] encode args error. [message = %+v, err = %v]", m, err)
		return
	}

	values := make([]reflect.Value, 2)
	values[0] = reflect.ValueOf(m.Session) // session
	values[1] = reflect.ValueOf(m.Args)    // args
	fi.Value.Call(values)
}

func InvokeRemoteFunc(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) {
	if app == nil {
		clog.Errorf("[InvokeRemoteFunc] app is nil. [message = %+v]", m)
		return
	}

	if err := EncodeRemoteArgs(app, fi, m); err != nil {
		clog.Errorf("[InvokeRemoteFunc] encode args error. [message = %+v, err = %v]", m, err)
		replyReponseCode(m, ccode.RPCRemoteExecuteError)
		return
	}

	values := make([]reflect.Value, fi.InArgsLen)
	if fi.InArgsLen > 0 {
		values[0] = reflect.ValueOf(m.Args) // args
	}

	if m.IsCluster {
		rets := fi.Value.Call(values)

		if m.Reply == "" {
			return
		}

		cutils.Try(func() {
			rsp := retValue(app, rets)
			replyResponse(m, rsp)

		}, func(errString string) {
			replyReponseCode(m, ccode.RPCRemoteExecuteError)
			clog.Errorf("[InvokeRemoteFunc] invoke error. [message = %+v, err = %s]", m, errString)
		})
	} else {
		cutils.Try(func() {
			if m.ChanResult == nil {
				fi.Value.Call(values)
			} else {
				rets := fi.Value.Call(values)
				rsp := retValue(app, rets)
				m.ChanResult <- rsp
			}
		}, func(errString string) {
			if m.ChanResult != nil {
				m.ChanResult <- nil
			}

			clog.Errorf("[InvokeRemoteFunc] invoke error.[source = %s, target = %s -> %s, funcType = %v, err = %+v]",
				m.Source,
				m.Target,
				m.FuncName,
				fi.InArgs,
				errString,
			)
		})
	}
}

func EncodeRemoteArgs(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) error {
	if m.IsCluster {
		if fi.InArgsLen == 0 || m.Args == nil {
			return nil
		}

		return EncodeArgs(app, fi, 0, m)
	}

	return nil
}

func EncodeLocalArgs(app cfacade.IApplication, fi *creflect.FuncInfo, m *cfacade.Message) error {
	return EncodeArgs(app, fi, 1, m)
}

func EncodeArgs(app cfacade.IApplication, fi *creflect.FuncInfo, index int, m *cfacade.Message) error {
	if index >= len(fi.InArgs) {
		return cerror.Errorf("encode args index out of bounds. [index = %d, len = %d]",
			index,
			len(fi.InArgs),
		)
	}

	argBytes, ok := m.Args.([]byte)
	if !ok {
		return cerror.Errorf("Encode args error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
	}

	argValue := reflect.New(fi.InArgs[index].Elem()).Interface()
	err := app.Serializer().Unmarshal(argBytes, argValue)
	if err != nil {
		return cerror.Errorf("Encode args unmarshal error.[source = %s, target = %s -> %s, funcType = %v]",
			m.Source,
			m.Target,
			m.FuncName,
			fi.InArgs,
		)
	}

	m.Args = argValue

	return nil
}

func retValue(app cfacade.IApplication, rets []reflect.Value) *cproto.Response {
	rsp := &cproto.Response{
		Code: ccode.OK,
	}

	retsLen := len(rets)
	switch retsLen {
	case 1:
		if val := rets[0].Interface(); val != nil {
			if c, ok := val.(int32); ok {
				rsp.Code = c
			}
		}
	case 2:
		if !rets[0].IsNil() {
			data, err := app.Serializer().Marshal(rets[0].Interface())
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

func replyReponseCode(m *cfacade.Message, errCode int32) {
	rsp := &cproto.Response{
		Code: errCode,
	}

	replyResponse(m, rsp)
}

func replyResponse(m *cfacade.Message, rsp *cproto.Response) {
	replyData, err := proto.Marshal(rsp)
	if err != nil {
		clog.Warnf("Source = %s, Target = %s,  err := %v", m.Source, m.Target, err)
		return
	}

	err = cnats.ReplySync(m.ReqID, m.Reply, replyData)
	if err != nil {
		clog.Warn(err)
	}

	m.Destory()
}
