package cherryConnector

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/net/session"
)

type SimpleComponentOptions struct {
	Connector      cherryInterfaces.IConnector
	Encode         cherryInterfaces.PacketEncoder
	Decode         cherryInterfaces.PacketDecoder
	Serializer     cherryInterfaces.ISerializer
	ForwardMessage bool
	Heartbeat      int
}

type SimpleComponent struct {
	cherryInterfaces.BaseComponent
	SimpleComponentOptions
	connCount      *Connection
	sessionService *cherrySession.SessionComponent
}

func (t *SimpleComponent) Name() string {
	return cherryConst.ConnectorSimpleComponent
}

func (t *SimpleComponent) Init() {
}

func (t *SimpleComponent) AfterInit() {
	t.sessionService = t.App().Find(cherryConst.SessionComponent).(*cherrySession.SessionComponent)
	if t.sessionService == nil {
		panic("please load session.SessionComponent")
	}
}
