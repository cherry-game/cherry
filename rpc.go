package cherry

import (
	"context"
	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cmsg "github.com/cherry-game/cherry/net/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	crouter "github.com/cherry-game/cherry/net/router"
	"github.com/golang/protobuf/proto"
	"reflect"
	"time"
)

func GetRPC() cfacade.RPCClient {
	return _clusterComponent.RPCClient
}

func RequestRemote(nodeId string, route string, arg proto.Message, reply proto.Message, timeout ...time.Duration) int32 {
	if reply != nil && reflect.TypeOf(reply).Kind() != reflect.Ptr {
		return ccode.RPCReplyParamsError
	}

	var requestTimeout time.Duration
	if len(timeout) > 0 {
		requestTimeout = timeout[0]
	}

	bytes, err := _thisApp.Marshal(arg)
	if err != nil {
		clog.Warnf("[RequestRemote] marshal error. [nodeId = %s, route = %s, err = %+v]",
			nodeId,
			route,
			err,
		)
		return ccode.RPCMarshalError
	}

	request := cproto.GetRequest()
	defer cproto.PutRequest(request)

	request.Route = route
	request.Data = bytes

	rsp, err := _clusterComponent.RequestRemote(nodeId, request, requestTimeout)
	if err != nil || ccode.IsFail(rsp.Code) {
		clog.Warnf("[RequestRemote] marshal error. [nodeId = %s, route = %s, err = %+v]",
			nodeId,
			route,
			err,
		)
		return rsp.Code
	}

	if err = proto.Unmarshal(rsp.Data, reply); err != nil {
		return ccode.RPCUnmarshalError
	}

	return ccode.OK
}

func RequestRemoteByRoute(route string, arg proto.Message, reply proto.Message, timeout ...time.Duration) int32 {
	rt, err := cmsg.DecodeRoute(route)
	if err != nil {
		clog.Warnf("[RPCByRoute] decode route fail.. [error = %s]", err)
		return ccode.RPCRouteDecodeError
	}

	member, err := crouter.Route(context.Background(), rt.NodeType(), nil)
	if err != nil {
		clog.Warnf("[RPCByRoute]get node router is fail. [route = %s] [error = %s]", route, err)
		return ccode.RPCRouteHashError
	}

	return RequestRemote(member.GetNodeId(), route, arg, reply, timeout...)
}

func PublishRemote(nodeId string, route string, arg proto.Message) {
	if nodeId == "" {
		decode, err := cmsg.DecodeRoute(route)
		if err != nil {
			clog.Warnf("[PublishRemote] decode route fail. [nodeId = %s, route = %s, val = %+v]",
				nodeId,
				route,
				arg,
			)
			return
		}

		member, err := crouter.Route(context.Background(), decode.NodeType(), nil)
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

	bytes, err := _thisApp.Marshal(arg)
	if err != nil {
		clog.Warnf("[PublishRemote] marshal error. [nodeId = %s, route = %s, err = %+v]",
			nodeId,
			route,
			err,
		)
		return
	}

	request := cproto.GetRequest()
	defer cproto.PutRequest(request)

	request.Route = route
	request.Data = bytes

	_clusterComponent.PublishRemote(nodeId, request)
}

func PublishRemoteByRoute(route string, arg proto.Message) {
	PublishRemote("", route, arg)
}
