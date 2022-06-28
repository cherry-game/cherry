package cherryCommand

import (
	cfacade "github.com/cherry-game/cherry/facade"
	cpacket "github.com/cherry-game/cherry/net/packet"
	csession "github.com/cherry-game/cherry/net/session"
)

// ICommand request data command for client
type ICommand interface {
	PacketType() cpacket.Type
	Do(session *csession.Session, packet cfacade.IPacket)
}
