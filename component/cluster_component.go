package cherryComponent

import (
	cherryProto "github.com/cherry-game/cherry/net/cluster/proto"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryRoute "github.com/cherry-game/cherry/net/route"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type IClusterComponent interface {
	SendUserMessage(session *cherrySession.Session, route *cherryRoute.Route, msg *cherryMessage.Message)
	SendSysMessage(nodeId string, msg *cherryProto.Message)
}
