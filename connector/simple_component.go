package cherryConnector

import (
	"github.com/phantacix/cherry/const"
	"github.com/phantacix/cherry/interfaces"
	"github.com/phantacix/cherry/session"
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
	return "connector.simple"
}

func (t *SimpleComponent) Init() {
}

func (t *SimpleComponent) AfterInit() {
	t.sessionService = t.App().Find(cherryConst.SessionComponent).(*cherrySession.SessionComponent)
	if t.sessionService == nil {
		panic("please load session.SessionComponent")
	}
}
