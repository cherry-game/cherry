package pomelo

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
)

const (
	ResponseFuncName = "response" // remote function name for response
	PushFuncName     = "push"     // remote function name for push
	KickFuncName     = "kick"     // remote function name for kick
	BroadcastName    = "broadcast" // remote function name for broadcast
)

// ActorBase provides convenience methods for actors to respond, push, kick,
// and broadcast messages to clients. Embed it in your actor to inherit these
// methods.
type ActorBase struct {
	cactor.Base
}

// Response sends a response payload back to the client for the given session.
func (p *ActorBase) Response(session *cproto.Session, v any) {
	Response(p, session.AgentPath, session.Sid, session.GetMID(), v)
}

// ResponseCode sends a status-code response back to the client for the given session.
func (p *ActorBase) ResponseCode(session *cproto.Session, statusCode int32) {
	ResponseCode(p, session.AgentPath, session.Sid, session.GetMID(), statusCode)
}

// Push sends a push message to the client identified by the session's sid.
func (p *ActorBase) Push(session *cproto.Session, route string, v any) {
	PushWithSID(p, session.AgentPath, session.Sid, route, v)
}

// PushWithUIDS sends a push message to clients matching the given uid list or to
// all connected clients when allUID is true.
func (p *ActorBase) PushWithUIDS(agentPath string, uidList []int64, allUID bool, route string, v interface{}) {
	PushWithUIDS(p, agentPath, uidList, allUID, route, v)
}

// Kick sends a kick message to the client identified by the session's sid.
func (p *ActorBase) Kick(session *cproto.Session, reason any, closed bool) {
	Kick(p, session.AgentPath, session.Sid, reason, closed)
}

// Response looks up the agent by request mid and sends a response payload back to the client.
func Response(iActor cfacade.IActor, agentPath, sid string, mid uint32, v any) {
	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Response] Marshal error. agentPath = %s, v = %+v", agentPath, v)
		return
	}

	rsp := &cproto.PomeloResponse{
		Sid:  sid,
		Mid:  mid,
		Data: data,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

// ResponseCode looks up the agent by request mid and sends a status-code response back to the client.
func ResponseCode(iActor cfacade.IActor, agentPath, sid string, mid uint32, statusCode int32) {
	rsp := &cproto.PomeloResponse{
		Sid:  sid,
		Mid:  mid,
		Code: statusCode,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

// Push looks up the agent by sid or uid and sends a push message to the client.
func Push(iActor cfacade.IActor, agentPath, sid string, uid cfacade.UID, route string, v any) {
	if sid == "" && uid < 1 {
		clog.Warnf("[Push] sid or uid value error. agentPath = %s, route = %s, sid = %s, uid = %d",
			agentPath,
			route,
			sid,
			uid,
		)
		return
	}

	if route == "" {
		clog.Warnf("[Push] route value error. agentPath = %s, route = %s", agentPath, route)
		return
	}

	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Push] Marshal error. agentPath = %s, route = %s, v = %+v", agentPath, route, v)
		return
	}

	rsp := &cproto.PomeloPush{
		Sid:   sid,
		Uid:   uid,
		Route: route,
		Data:  data,
	}

	iActor.Call(agentPath, PushFuncName, rsp)
}

// PushWithSID looks up the agent by session id and sends a push message.
func PushWithSID(iActor cfacade.IActor, agentPath, sid, route string, v any) {
	Push(iActor, agentPath, sid, 0, route, v)
}

// PushWithUID looks up the agent by user id and sends a push message.
func PushWithUID(iActor cfacade.IActor, agentPath string, uid cfacade.UID, route string, v any) {
	Push(iActor, agentPath, "", uid, route, v)
}

// PushWithUIDS sends a push message to agents matching the uid list or to all connected
// clients when allUID is true.
func PushWithUIDS(iActor cfacade.IActor, agentPath string, uidList []int64, allUID bool, route string, v any) {
	if !allUID && len(uidList) < 1 {
		clog.Warnf("[PushWithUIDS] uidList value error. agentPath = %s, route = %s", agentPath, route)
		return
	}

	if route == "" {
		clog.Warnf("[PushWithUIDS] route value error. agentPath = %s, route = %s", agentPath, route)
		return
	}

	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[PushWithUIDS] Marshal error. agentPath = %s, route = %s, v = %+v", agentPath, route, v)
		return
	}

	rsp := &cproto.PomeloBroadcast{
		Route: route,
		Data:  data,
	}

	if allUID {
		rsp.PushType = cproto.PomeloBroadcast_AllUID
	} else {
		rsp.PushType = cproto.PomeloBroadcast_UID
		rsp.UidList = uidList
	}

	iActor.Call(agentPath, BroadcastName, rsp)
}

// Kick looks up the agent by session id and sends a kick message to the client.
func Kick(iActor cfacade.IActor, agentPath, sid string, reason any, closed bool) {
	data, err := iActor.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. agentPath = %s, sid = %s, reason = %+v", agentPath, sid, reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Sid:    sid,
		Reason: data,
		Close:  closed,
	}

	iActor.Call(agentPath, KickFuncName, rsp)
}

// KickUID looks up the agent by user id and sends a kick message to the client.
func KickUID(iActor cfacade.IActor, agentPath string, uid cfacade.UID, reason any, closed bool) {
	data, err := iActor.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[KickUID] Marshal error. agentPath = %s, uid = %d, reason = %+v", agentPath, uid, reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Uid:    uid,
		Reason: data,
		Close:  closed,
	}

	iActor.Call(agentPath, KickFuncName, rsp)
}
