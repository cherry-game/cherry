package pomelo

import (
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
	data, err := p.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Response] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloResponse{
		Sid:  session.Sid,
		Mid:  session.Mid,
		Data: data,
	}

	p.Call(session.AgentPath, ResponseFuncName, rsp)
}

func (p *ActorBase) ResponseCode(session *cproto.Session, statusCode int32) {
	rsp := &cproto.PomeloResponse{
		Sid:  session.Sid,
		Mid:  session.Mid,
		Code: statusCode,
	}

	p.Call(session.AgentPath, ResponseFuncName, rsp)
}

func (p *ActorBase) Push(session *cproto.Session, route string, v interface{}) {
	if route == "" {
		clog.Warn("[Push] route value error.")
		return
	}

	data, err := p.App().Serializer().Marshal(v)
	if err != nil {
		clog.Warnf("[Push] Marshal error. v = %+v", v)
		return
	}

	rsp := &cproto.PomeloPush{
		Sid:   session.Sid,
		Route: route,
		Data:  data,
	}

	p.Call(session.AgentPath, PushFuncName, rsp)
}

func (p *ActorBase) Kick(session *cproto.Session, reason interface{}, close bool) {
	data, err := p.App().Serializer().Marshal(reason)
	if err != nil {
		clog.Warnf("[Push] Marshal error. reason = %+v", reason)
		return
	}

	rsp := &cproto.PomeloKick{
		Sid:    session.Sid,
		Reason: data,
		Close:  close,
	}

	p.Call(session.AgentPath, KickFuncName, rsp)
}
