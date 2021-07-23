package cherryComponent

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type IHandlerComponent interface {
	PostEvent(event facade.IEvent)
	PostMessage(session *cherrySession.Session, msg *cherryMessage.Message)
}
