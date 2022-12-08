package cherry

import (
	"context"
	ccode "github.com/cherry-game/cherry/code"
	cerror "github.com/cherry-game/cherry/error"
	creflect "github.com/cherry-game/cherry/extend/reflect"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cmsg "github.com/cherry-game/cherry/net/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	crouter "github.com/cherry-game/cherry/net/router"
	"time"
)

func GetRPC() cfacade.RPCClient {
	return _clusterComponent.RPCClient
}

func RequestRemote(nodeId string, route string, arg, reply interface{}, timeout ...time.Duration) int32 {
	if arg != nil && creflect.IsNotPtr(arg) {
		clog.Warnf("[RequestRemote] arg is not ptr. [nodeId = %s, route = %s, arg = %+v]",
			nodeId,
			route,
			arg,
		)
		return ccode.RPCReplyParamsError
	}

	if reply != nil && creflect.IsNotPtr(reply) {
		clog.Warnf("[RequestRemote] reply is not ptr. [nodeId = %s, route = %s, reply = %+v]",
			nodeId,
			route,
			reply,
		)
		return ccode.RPCReplyParamsError
	}

	var requestTimeout time.Duration
	if len(timeout) > 0 {
		requestTimeout = timeout[0]
	}

	var bytes []byte
	var err error
	if arg != nil {
		bytes, err = _thisApp.Marshal(arg)
		if err != nil {
			clog.Warnf("[RequestRemote] arg marshal error. [nodeId = %s, route = %s, err = %+v]",
				nodeId,
				route,
				err,
			)
			return ccode.RPCMarshalError
		}
	}

	request := cproto.GetRequest()
	defer request.Recycle()

	request.Route = route
	request.Data = bytes

	rsp, err := _clusterComponent.RequestRemote(nodeId, request, requestTimeout)
	if err != nil {
		clog.Warnf("[RequestRemote] response error. [nodeId = %s, route = %s, err = %+v]",
			nodeId,
			route,
			err,
		)
		return rsp.Code
	}

	if ccode.IsFail(rsp.Code) {
		return rsp.Code
	}

	if reply != nil && rsp.Data != nil {
		if err = _thisApp.Unmarshal(rsp.Data, reply); err != nil {
			clog.Warnf("[RequestRemote] reply unmarshal error. [nodeId = %s, route = %s, data = %+v, err = %+v]",
				nodeId,
				route,
				rsp.Data,
				err,
			)
			return ccode.RPCUnmarshalError
		}
	}

	return rsp.Code
}

func RequestRemoteByRoute(route string, arg, reply interface{}, timeout ...time.Duration) int32 {
	rt, err := cmsg.DecodeRoute(route)
	if err != nil {
		clog.Warnf("[RequestRemoteByRoute] decode route fail.. [route = %s, arg = %s, reply = %+v, error = %s]",
			route,
			arg,
			reply,
			err,
		)
		return ccode.RPCRouteDecodeError
	}

	member, err := crouter.Route(context.Background(), rt, nil)
	if err != nil {
		clog.Warnf("[RequestRemoteByRoute] get node router is fail. [route = %s] [error = %s]", route, err)
		return ccode.RPCRouteHashError
	}

	return RequestRemote(member.GetNodeId(), route, arg, reply, timeout...)
}

func PublishRemote(nodeId string, route string, arg interface{}) {
	if arg != nil && creflect.IsNotPtr(arg) {
		clog.Warnf("[PublishRemote] arg is not ptr. [nodeId = %s, route = %s, arg = %+v]",
			nodeId,
			route,
			arg,
		)
		return
	}

	if nodeId == "" {
		rt, err := cmsg.DecodeRoute(route)
		if err != nil {
			clog.Warnf("[PublishRemote] decode route fail. [nodeId = %s, route = %s, arg = %+v]",
				nodeId,
				route,
				arg,
			)
			return
		}

		member, err := crouter.Route(context.Background(), rt, nil)
		if err != nil {
			clog.Warnf("[PublishRemote] get node router fail. [nodeId = %s, route = %s, err = %+v]",
				nodeId,
				route,
				err,
			)
			return
		}

		nodeId = member.GetNodeId()
	}

	if route == "" {
		clog.Warnf("[PublishRemote] route is nil. [nodeId = %s, route = %s, val = %+v]",
			nodeId,
			route,
			arg,
		)
		return
	}

	var bytes []byte
	var err error
	if arg != nil {
		bytes, err = _thisApp.Marshal(arg)
		if err != nil {
			clog.Warnf("[PublishRemote] arg marshal error. [nodeId = %s, route = %s, err = %+v]",
				nodeId,
				route,
				err,
			)
			return
		}
	}

	request := cproto.GetRequest()
	defer request.Recycle()

	request.Route = route
	request.Data = bytes

	_clusterComponent.PublishRemote(nodeId, request)
}

func PublishRemoteByRoute(route string, arg interface{}) {
	PublishRemote("", route, arg)
}

func Kick(nodeId string, uid cfacade.UID, val interface{}, close bool) error {
	if creflect.IsNotPtr(val) {
		return cerror.Errorf("[kick] val is not ptr. [nodeId = %s, uid = %s, val = %+v]",
			nodeId,
			uid,
			val,
		)
	}

	bytes, err := _thisApp.Marshal(val)
	if err != nil {
		return err
	}

	kick := &cproto.Kick{
		Uid:   uid,
		Data:  bytes,
		Close: close,
	}

	return _clusterComponent.PublishKick(nodeId, kick)
}

func Push(frontendId string, route string, uid cfacade.UID, val interface{}) error {
	if creflect.IsNotPtr(val) {
		return cerror.Errorf("[Push] val is not ptr. [frontendId = %s, route = %s, uid = %s, val = %+v]",
			frontendId,
			route,
			uid,
			val,
		)
	}

	bytes, err := _thisApp.Marshal(val)
	if err != nil {
		return err
	}

	push := &cproto.Push{
		Route: route,
		Uid:   uid,
		Data:  bytes,
	}

	return _clusterComponent.PublishPush(frontendId, push)
}
