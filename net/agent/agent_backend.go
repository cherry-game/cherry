package cherryAgent

import (
	"context"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryProto "github.com/cherry-game/cherry/net/proto"
	cherryRouter "github.com/cherry-game/cherry/net/router"
	cherrySession "github.com/cherry-game/cherry/net/session"
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
}

func (a *AgentBackend) Kick(reason interface{}) {
	a.rpcClient.SendKick(a.session.FrontendId(), a.session.UID(), reason)
}

func (a *AgentBackend) Response(mid uint, val interface{}, isError ...bool) {
	if mid < 1 {
		a.session.Warnf("[Response] mid error. [mid = %v]", mid)
		return
	}

	data, err := a.Marshal(val)
	if err != nil {
		a.session.Warnf("[Response] marshal data error. [val = %v]", val)
		return
	}

	errFlag := false
	if len(isError) > 0 {
		errFlag = isError[0]
	}

	localPacket := &cherryProto.LocalPacket{
		Session: &cherryProto.Session{
			Uid:        a.session.UID(),
			FrontendId: a.session.FrontendId(),
		},
		MsgType: int32(cherryMessage.Response),
		MsgId:   uint32(mid),
		IsError: errFlag,
		Data:    data,
	}

	err = a.rpcClient.CallLocal(a.session.FrontendId(), localPacket)
	if err != nil {
		a.session.Warnf("[Response] return error.. [val = %v] [err = %v]", val, err)
	}
}

func (a *AgentBackend) SendRaw(_ []byte) {
	a.session.Warnf("[SendRaw] %v", cherryError.ClusterNoImplement)
}

func (a *AgentBackend) RPC(route string, val interface{}, rsp *cherryProto.Response) {
	decode, err := cherryMessage.DecodeRoute(route)
	if err != nil {
		cherryLogger.Warnf("[RPC] decode route fail. [route = %s] [error = %v]", route, err)
		rsp.Code = cherryCode.RPCRouteDecodeError
		return
	}

	ctx := context.WithValue(context.Background(), cherryConst.SessionKey, a.session)
	member, err := cherryRouter.Route(ctx, decode.NodeType(), nil)
	if err != nil {
		cherryLogger.Warnf("get node router is fail. [error = %s]", err)
		rsp.Code = cherryCode.RPCRouteHashError
		return
	}

	a.rpcClient.CallRemote(member.GetNodeId(), route, val, 0, rsp)
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
