package main

import (
	"context"
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test_sample1/constant"
	cherryConst "github.com/cherry-game/cherry/const"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryDiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

func main() {
	gateApp := cherry.Configure("../../config/", "sample1", "gate-1")
	cherry.SetSerializer(cherrySerializer.NewJSON())

	cherry.AddNodeRouter(constant.GameNodeType, gameNodeRoute)
	cherry.AddNodeRouter(constant.CrossNodeType, crossNodeRoute)

	//cherry.RegisterComponent(&MockClientComponent{})

	cherry.RegisterHandler(&UserHandler{})
	cherry.SetHandlerOptions(cherryHandler.WithPrintRouteLog(true))

	cherry.RegisterConnector(cherryConnector.NewWS(gateApp.Address()))

	cherry.Run(true, cherry.Cluster)
}

func gameNodeRoute(ctx context.Context, _ string, _ *cherryMessage.Message) (cherryFacade.IMember, error) {
	session, found := ctx.Value(cherryConst.SessionKey).(*cherrySession.Session)
	if found == false {
		return nil, cherryError.SessionNotFoundInContext
	}

	serverId := session.GetString(constant.GameServerId)

	member, found := cherryDiscovery.GetMember(serverId)
	if found {
		return member, nil
	}

	return nil, cherryError.DiscoveryGetMemberListIsEmpty
}

func crossNodeRoute(ctx context.Context, _ string, _ *cherryMessage.Message) (cherryFacade.IMember, error) {
	session, found := ctx.Value(cherryConst.SessionKey).(*cherrySession.Session)
	if found == false {
		return nil, cherryError.SessionNotFoundInContext
	}

	serverId := session.GetString("cross_id")
	member, found := cherryDiscovery.GetMember(serverId)
	if found {
		return member, nil
	}

	return nil, cherryError.DiscoveryGetMemberListIsEmpty
}
