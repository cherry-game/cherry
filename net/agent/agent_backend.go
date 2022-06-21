package cherryAgent

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryRouter "github.com/cherry-game/cherry/net/router"
	cherrySession "github.com/cherry-game/cherry/net/session"
	cherryProfile "github.com/cherry-game/cherry/profile"
)

type AgentBackend struct {
	cherryFacade.IApplication
	rpcClient cherryFacade.RPCClient
	session   *cherrySession.Session
	ip        string
}

func NewAgentBackend(app cherryFacade.IApplication, rpcClient cherryFacade.RPCClient, ip string) AgentBackend {
	return AgentBackend{
		IApplication: app,
		rpcClient:    rpcClient,
		ip:           ip,
	}
}

func (a *AgentBackend) SetSession(session *cherrySession.Session) {
	a.session = session
}

func (a *AgentBackend) Push(route string, val interface{}) {
	a.rpcClient.SendPush(a.session.FrontendId(), route, a.session.UID(), val)

	if cherryProfile.Debug() {
		a.session.Debugf("[Push] ok. [frontend = %s, route = %s, val = %+v]",
			a.session.FrontendId(),
			route,
			val,
		)
	}
}

func (a *AgentBackend) Kick(reason interface{}) {
	a.rpcClient.SendKick(a.session.FrontendId(), a.session.UID(), reason)

	if cherryProfile.Debug() {
		a.session.Debugf("[Kick] ok. [frontend = %s, reason = %+v]",
			a.session.FrontendId(),
			reason,
		)
	}
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

	data, err := a.Marshal(val)
	if err != nil {
		a.session.Errorf("[Response] marshal error. [frontend = %s, mid = %d, isErr = %v, val = %+v]",
			a.session.FrontendId(),
			mid,
			isErr,
			val,
		)
		return
	}

	localPacket := &cherryProto.LocalPacket{
		Session: &cherryProto.Session{
			Uid:        a.session.UID(),
			FrontendId: a.session.FrontendId(),
		},
		MsgType: int32(cherryMessage.Response),
		MsgId:   uint32(mid),
		IsError: isErr,
		Data:    data,
	}

	err = a.rpcClient.CallLocal(a.session.FrontendId(), localPacket)
	if err != nil {
		a.session.Errorf("[Response] call error. [frontend = %s, mid = %d, isErr = %v, val = %+v]",
			a.session.FrontendId(),
			mid,
			isErr,
			val,
		)
	}

	if cherryProfile.Debug() {
		a.session.Debugf("[Response] ok. [frontend = %s, mid = %d, err = %v, val = %+v]",
			a.session.FrontendId(),
			mid,
			isErr,
			val,
		)
	}
}

func (a *AgentBackend) SendRaw(_ []byte) {
	a.session.Errorf("[SendRaw] %v", cherryError.ClusterNoImplement)
}

func (a *AgentBackend) RPC(route string, val interface{}, rsp *cherryProto.Response) {
	decode, err := cherryMessage.DecodeRoute(route)
	if err != nil {
		a.session.Warnf("[RPC] decode route fail. [route = %s, err = %v, val = %+v]",
			route,
			err,
			val,
		)
		rsp.Code = cherryCode.RPCRouteDecodeError
		return
	}

	ctx := context.WithValue(context.Background(), cherryConst.SessionKey, a.session)
	member, err := cherryRouter.Route(ctx, decode.NodeType(), nil)
	if err != nil {
		a.session.Warnf("[RPC] get router fail. [route = %s, err = %v, val = %+v]",
			route,
			err,
			val,
		)
		rsp.Code = cherryCode.RPCRouteHashError
		return
	}

	a.rpcClient.CallRemote(member.GetNodeId(), route, val, 0, rsp)

	if cherryProfile.Debug() {
		a.session.Debugf("[RPC] ok. [node = %s, route = %s, err = %v, val = %+v]",
			member.GetNodeId(),
			route,
			err,
			val,
		)
	}
}

func (a *AgentBackend) Close() {
	if a.session.State() == cherrySession.Closed {
		return
	}

	a.session.SetState(cherrySession.Closed)
	a.session.OnCloseListener()
}

func (a *AgentBackend) RemoteAddr() string {
	return a.ip
}
