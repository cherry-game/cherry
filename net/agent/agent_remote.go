package cherryAgent

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"net"
)

type AgentRemote struct {
	sid        cherryFacade.SID
	gateClient interface{}

	Session    *cherrySession.Session
	frontendId cherryFacade.FrontendId
	rpcClient  func()
}

func (a *AgentRemote) Push(route string, val interface{}) error {
	return nil
}

func (a *AgentRemote) Response(mid uint, val interface{}, isError ...bool) error {
	return nil
}

func (a *AgentRemote) Kick(reason interface{}) error {
	return nil
}

func (a *AgentRemote) SendRaw(bytes []byte) error {
	return nil
}

func (a *AgentRemote) RPC(route string, val interface{}) error {
	return nil
}

func (a *AgentRemote) Close() {

}

func (a *AgentRemote) RemoteAddr() net.Addr {
	return nil
}
