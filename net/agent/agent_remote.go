package cherryAgent

import (
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

type AgentRemote struct {
	sid        cherryFacade.SID
	gateClient interface{}

	Session       *cherrySession.Session
	PacketDecoder cherryPacket.Decoder     // binary packet decoder
	PacketEncoder cherryPacket.Encoder     // binary packet encoder
	Serializer    cherryFacade.ISerializer // data serializer
	frontendId    cherryFacade.FrontendId
	rpcClient     func()
}
