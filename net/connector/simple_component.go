package cherryConnector

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/net/packet"
	"github.com/cherry-game/cherry/net/session"
)

type SimpleComponentOptions struct {
	Connector      cherryFacade.IConnector
	Encode         cherryPacket.Encoder
	Decode         cherryPacket.Decoder
	Serializer     cherryFacade.ISerializer
	ForwardMessage bool
	Heartbeat      int
}

type SimpleComponent struct {
	cherryFacade.Component
	SimpleComponentOptions
	connCount        *ConnectStat
	sessionComponent *cherrySession.SessionComponent
}

func (t *SimpleComponent) Name() string {
	return cherryConst.ConnectorSimpleComponent
}

func (t *SimpleComponent) Init() {
}

func (t *SimpleComponent) OnAfterInit() {
	t.sessionComponent = t.App().Find(cherryConst.SessionComponent).(*cherrySession.SessionComponent)
	if t.sessionComponent == nil {
		panic("please load session.SessionComponent")
	}
}
