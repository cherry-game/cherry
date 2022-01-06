package cherryRouter

import (
	"context"
	cherryError "github.com/cherry-game/cherry/error"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryDiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	"math/rand"
	"time"
)

var (
	routesMap = make(map[string]RoutingFunc)
)

type RoutingFunc func(
	ctx context.Context,
	nodeType string,
	msg *cherryMessage.Message,
) (cherryFacade.IMember, error)

func randRoute(nodeType string) (cherryFacade.IMember, error) {
	s := rand.NewSource(time.Now().Unix())
	rnd := rand.New(s)

	memberList := cherryDiscovery.ListByType(nodeType)
	if len(memberList) < 1 {
		return nil, cherryError.DiscoveryGetMemberListIsEmpty
	}

	if len(memberList) == 1 {
		return memberList[0], nil
	}

	server := memberList[rnd.Intn(len(memberList))]
	return server, nil
}

func AddRoute(nodeType string, routingFunction RoutingFunc) {
	if _, ok := routesMap[nodeType]; ok {
		cherryLogger.Warnf("overriding the route to [nodeType = %s]", nodeType)
	}

	routesMap[nodeType] = routingFunction
}

func Route(ctx context.Context, nodeType string, msg *cherryMessage.Message) (cherryFacade.IMember, error) {
	routeFunc, ok := routesMap[nodeType]
	if !ok {
		return randRoute(nodeType)
	}

	return routeFunc(ctx, nodeType, msg)
}
