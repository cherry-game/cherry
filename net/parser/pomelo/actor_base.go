package pomelo

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	cproto "github.com/cherry-game/cherry/net/proto"
)

const (
	ResponseFuncName = "response"
	PushFuncName     = "push"
	KickFuncName     = "kick"
	BroadcastName    = "broadcast"
)

type ActorBase struct {
	cactor.Base
}

func (p *ActorBase) Response(session *cproto.Session, v any) {
	Response(p, session.AgentPath, session.Sid, session.GetMID(), v)
}

func (p *ActorBase) ResponseCode(session *cproto.Session, statusCode int32) {
	ResponseCode(p, session.AgentPath, session.Sid, session.GetMID(), statusCode)
}

func (p *ActorBase) Push(session *cproto.Session, route string, v any) {
	PushWithSID(p, session.AgentPath, session.Sid, route, v)
}

func (p *ActorBase) PushWithUIDS(agentPath string, uidList []int64, allUID bool, route string, v interface{}) {
	PushWithUIDS(p, agentPath, uidList, allUID, route, v)
}

func (p *ActorBase) Kick(session *cproto.Session, reason any, closed bool) {
	Kick(p, session.AgentPath, session.Sid, reason, closed)
}

// 根据request的mid找到agent，返回消息给客户端
func Response(iActor cfacade.IActor, agentPath, sid string, mid uint32, v any) {
	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Response] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloResponse{
		Sid:  sid,
		Mid:  mid,
		Data: data,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

// 根据request的mid找到agent，返回消息给客户端
func ResponseCode(iActor cfacade.IActor, agentPath, sid string, mid uint32, statusCode int32) {
	rsp := &cproto.PomeloResponse{
		Sid:  sid,
		Mid:  mid,
		Code: statusCode,
	}

	iActor.Call(agentPath, ResponseFuncName, rsp)
}

// 根据sid或uid找到agent，推送消息给客户端
func Push(iActor cfacade.IActor, agentPath, sid string, uid cfacade.UID, route string, v any) {
	if sid == "" && uid < 1 {
		clog.Warn("[Push] sid or uid value error.")
		return
	}

	if route == "" {
		clog.Warn("[Push] route value error.")
		return
	}

	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Push] Marshal error. route =%s, v = %+v", route, v)
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

// 根据sid找到agent，推送消息给客户端
func PushWithSID(iActor cfacade.IActor, agentPath, sid, route string, v any) {
	Push(iActor, agentPath, sid, 0, route, v)
}

// 根据uid找到agent，推送消息给客户端
func PushWithUID(iActor cfacade.IActor, agentPath string, uid cfacade.UID, route string, v any) {
	Push(iActor, agentPath, "", uid, route, v)
}

// 根据uidList或allUID匹配找到Agent，下发数据给客户端
func PushWithUIDS(iActor cfacade.IActor, agentPath string, uidList []int64, allUID bool, route string, v any) {
	if !allUID && len(uidList) < 1 {
		clog.Warn("[Broadcast] uidList value error.")
		return
	}

	if route == "" {
		clog.Warn("[Broadcast] route value error.")
		return
	}

	data, err := iActor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloBroadcast{
		Route: route,
		Data:  data,
	}

	if allUID {
		rsp.PushType = cproto.PomeloBroadcast_AllUID
	} else {
		rsp.UidList = uidList
	}

	iActor.Call(agentPath, BroadcastName, rsp)
}

// 根据sid找到agent，下发踢除消息给客户端
func Kick(iActor cfacade.IActor, agentPath, sid string, reason any, closed bool) {
	data, err := iActor.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. reason = %+v", reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Sid:    sid,
		Reason: data,
		Close:  closed,
	}

	iActor.Call(agentPath, KickFuncName, rsp)
}

func KickUID(iActor cfacade.IActor, agentPath string, uid cfacade.UID, reason any, closed bool) {
	data, err := iActor.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. reason = %+v", reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Uid:    uid,
		Reason: data,
		Close:  closed,
	}

	iActor.Call(agentPath, KickFuncName, rsp)
}
