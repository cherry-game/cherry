package cherryAgent

import cherryFacade "github.com/cherry-game/cherry/facade"

type AgentRemote struct {
	sid        cherryFacade.SID
	gateClient interface{}
}
