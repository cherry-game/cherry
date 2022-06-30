package main

import (
	"context"
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test_sample1/constant"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cdiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cconnector "github.com/cherry-game/cherry/net/connector"
	chandler "github.com/cherry-game/cherry/net/handler"
	cmsg "github.com/cherry-game/cherry/net/message"
	cserializer "github.com/cherry-game/cherry/net/serializer"
	csession "github.com/cherry-game/cherry/net/session"
)

func main() {
	gateApp := cherry.Configure("../../config/", "sample1", "gate-1")
	cherry.SetSerializer(cserializer.NewJSON())

	cherry.AddNodeRouter(constant.GameNodeType, gameNodeRoute)
	cherry.AddNodeRouter(constant.CrossNodeType, crossNodeRoute)

	//cherry.RegisterComponent(&MockClientComponent{})

	cherry.RegisterHandler(&UserHandler{})
	cherry.SetHandlerOptions(chandler.WithPrintRouteLog(true))

	cherry.RegisterConnector(cconnector.NewWS(gateApp.Address()))

	cherry.Run(true, cherry.Cluster)
}

func gameNodeRoute(_ context.Context, _ *cmsg.Route, session *csession.Session) (cfacade.IMember, error) {
	if session == nil || session.IsBind() == false {
		clog.Warnf("session is not bind.")
		return nil, nil
	}

	serverId := session.GetString(constant.GameServerId)
	if member, found := cdiscovery.GetMember(serverId); found {
		return member, nil
	}

	return nil, cerr.DiscoveryMemberListIsEmpty
}

func crossNodeRoute(_ context.Context, _ *cmsg.Route, session *csession.Session) (cfacade.IMember, error) {
	if session == nil || session.IsBind() == false {
		clog.Warnf("session is not bind.")
		return nil, nil
	}

	serverId := session.GetString("cross_id")
	if member, found := cdiscovery.GetMember(serverId); found {
		return member, nil
	}

	return nil, cerr.DiscoveryMemberListIsEmpty
}
