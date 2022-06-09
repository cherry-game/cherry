package cherryHandler

import (
	cherryLogger "github.com/cherry-game/cherry/logger"
	cp "github.com/cherry-game/cherry/net/proto"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type SessionHandler struct {
	Handler
}

func (p *SessionHandler) Name() string {
	return "sessionHandler"
}

func (p *SessionHandler) OnInit() {
	p.AddRemote("kick", p.kick)
	p.AddRemote("push", p.push)
}

func (p *SessionHandler) kick(request *cp.KickRequest) {
	session, found := cherrySession.GetByUID(request.Uid)
	if found == false {
		cherryLogger.Warnf("[kick] uid = %d not found.", request.Uid)
		return
	}
	session.Kick(request.Reason, true)
}

func (p *SessionHandler) push(request *cp.PushRequest) {
	session, found := cherrySession.GetByUID(request.Uid)
	if found == false {
		cherryLogger.Warnf("[push route = %s] uid = %d not found.", request.Route, request.Uid)
		return
	}

	session.Push(request.Route, request.Data)
}
