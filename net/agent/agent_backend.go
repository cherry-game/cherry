package cherryAgent

import (
	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cherryDiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cmsg "github.com/cherry-game/cherry/net/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	csession "github.com/cherry-game/cherry/net/session"
	"github.com/golang/protobuf/proto"
)

type AgentBackend struct {
	cfacade.IApplication
	rpcClient cfacade.RPCClient
	session   *csession.Session
	ip        string
}

func NewAgentBackend(app cfacade.IApplication, rpcClient cfacade.RPCClient, ip string) AgentBackend {
	return AgentBackend{
		IApplication: app,
		rpcClient:    rpcClient,
		ip:           ip,
	}
}

func (a *AgentBackend) SetSession(session *csession.Session) {
	a.session = session
}

func (a *AgentBackend) Push(route string, val interface{}) {
	bytes, err := a.Marshal(val)
	if err != nil {
		clog.Warnf("[Push] marshal error.  [frontend = %s, route = %s, val = %+v]",
			a.session.FrontendId(),
			route,
			val,
		)
	}

	push := &cproto.Push{
		Route: route,
		Uid:   a.session.UID(),
		Data:  bytes,
	}

	a.rpcClient.PublishPush(a.session.FrontendId(), push)
}

func (a *AgentBackend) Kick(reason interface{}) {
	bytes, err := a.Marshal(reason)
	if err != nil {
		clog.Warnf("[Kick] marshal error.  [frontend = %s, val = %+v, err = %s]",
			a.session.FrontendId(),
			reason,
			err,
		)
	}

	kick := &cproto.Kick{
		Uid:  a.session.UID(),
		Data: bytes,
	}

	nodeType, err := cherryDiscovery.GetType(a.session.FrontendId())
	if err != nil {
		clog.Warnf("[Kick] get node type error.  [frontend = %s, val = %+v, err = %s]",
			a.session.FrontendId(),
			reason,
			err,
		)
	}

	a.rpcClient.PublishKick(nodeType, kick)
}

func (a *AgentBackend) Response(mid uint, val interface{}, isError ...bool) {
	isErr := false
	if len(isError) > 0 {
		isErr = isError[0]
	}

	if mid < 1 {
		a.session.Errorf("[Response] mid error. [frontend = %s, mid = %d, isErr = %v, val = %+v]",
			a.session.FrontendId(),
			mid,
			isErr,
			val,
		)
		return
	}

	bytes, err := a.Marshal(val)
	if err != nil {
		a.session.Errorf("[Response] marshal error. [frontend = %s, mid = %d, isErr = %v, val = %+v]",
			a.session.FrontendId(),
			mid,
			isErr,
			val,
		)
		return
	}

	request := cproto.GetRequest()
	defer request.Recycle()

	request.Uid = a.session.UID()
	request.FrontendId = a.session.FrontendId()
	request.MsgType = int32(cmsg.Response)
	request.MsgId = uint32(mid)
	request.IsError = isErr
	request.Data = bytes

	a.rpcClient.PublishLocal(a.session.FrontendId(), request)
	if err != nil {
		a.session.Errorf("[Response] call error. [frontend = %s, mid = %d, isErr = %v, val = %+v]",
			a.session.FrontendId(),
			mid,
			isErr,
			val,
		)
	}
}

func (a *AgentBackend) RPC(nodeId string, route string, req proto.Message, rsp proto.Message) int32 {
	if _, err := cmsg.DecodeRoute(route); err != nil {
		a.session.Errorf("[RPC] decode route fail. [nodeId = %s, route = %s, req = %+v, err = %v]",
			nodeId,
			route,
			req,
			err,
		)
		return ccode.RPCRouteDecodeError
	}

	bytes, err := a.Marshal(req)
	if err != nil {
		a.session.Errorf("[Response] marshal error. [nodeId = %s, route = %s, req = %+v, err = %v]",
			nodeId,
			route,
			req,
			err,
		)
		return ccode.RPCMarshalError
	}

	request := cproto.GetRequest()
	defer request.Recycle()

	request.Route = route
	request.Uid = a.session.UID()
	request.Data = bytes

	rspData, err := a.rpcClient.RequestRemote(nodeId, request)
	if err != nil {
		a.session.Errorf("[Response] error. [nodeId = %s, route = %s, req = %+v, rsp = %+v, err = %v]",
			nodeId,
			route,
			req,
			rspData,
			err,
		)
		return ccode.RPCRemoteExecuteError
	}

	if ccode.IsFail(rspData.Code) {
		return rspData.Code
	}

	err = a.Unmarshal(rspData.Data, rsp)
	if err != nil {
		return ccode.RPCUnmarshalError
	}

	return ccode.OK
}

func (a *AgentBackend) SendRaw(data []byte) {
	a.rpcClient.Publish(a.session.FrontendId(), data)
}

func (a *AgentBackend) RemoteAddr() string {
	return a.ip
}

func (a *AgentBackend) Close() {
	if a.session.State() == csession.Closed {
		return
	}

	a.session.SetState(csession.Closed)
	a.session.OnCloseListener()
}
