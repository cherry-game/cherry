package cherryCommand

import (
	facade "github.com/cherry-game/cherry/facade"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

// ICommand request data command for client
type ICommand interface {
	GetType() cherryPacket.Type
	Do(session *cherrySession.Session, packet facade.IPacket)
}
