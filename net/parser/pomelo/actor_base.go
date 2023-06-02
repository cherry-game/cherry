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
)

type ActorBase struct {
	cactor.Base
}

func (p *ActorBase) Response(session *cproto.Session, v interface{}) {
	Response(p, session, v)
}

func (p *ActorBase) ResponseCode(session *cproto.Session, statusCode int32) {
	ResponseCode(p, session, statusCode)
}

func (p *ActorBase) Push(session *cproto.Session, route string, v interface{}) {
	Push(p, session, route, v)
}

func (p *ActorBase) Kick(session *cproto.Session, reason interface{}, close bool) {
	Kick(p, session, reason, close)
}

func Response(actor cfacade.IActor, session *cproto.Session, v interface{}) {
	data, err := actor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Response] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloResponse{
		Sid:  session.Sid,
		Mid:  session.Mid,
		Data: data,
	}

	actor.Call(session.AgentPath, ResponseFuncName, rsp)
}

func ResponseCode(actor cfacade.IActor, session *cproto.Session, statusCode int32) {
	rsp := &cproto.PomeloResponse{
		Sid:  session.Sid,
		Mid:  session.Mid,
		Code: statusCode,
	}

	actor.Call(session.AgentPath, ResponseFuncName, rsp)
}

func Push(actor cfacade.IActor, session *cproto.Session, route string, v interface{}) {
	if route == "" {
		clog.Warn("[Push] route value error.")
		return
	}

	data, err := actor.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Push] Marshal error. route =%s, v = %+v", route, v)
		return
	}

	rsp := &cproto.PomeloPush{
		Sid:   session.Sid,
		Route: route,
		Data:  data,
	}

	actor.Call(session.AgentPath, PushFuncName, rsp)
}

func Kick(actor cfacade.IActor, session *cproto.Session, reason interface{}, close bool) {
	data, err := actor.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[Kick] Marshal error. reason = %+v", reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Sid:    session.Sid,
		Reason: data,
		Close:  close,
	}

	actor.Call(session.AgentPath, KickFuncName, rsp)
}
