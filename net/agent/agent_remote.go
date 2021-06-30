package cherryAgent

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type AgentRemote struct {
	sid        cherryFacade.SID
	gateClient interface{}

	Session    *cherrySession.Session
	frontendId cherryFacade.FrontendId
	rpcClient  func()
}
