package cherryRouter

import (
	"context"
	cerr "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cdiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cmsg "github.com/cherry-game/cherry/net/message"
	"math/rand"
	"time"
)

var (
	routesMap = make(map[string]RoutingFunc)
)

type RoutingFunc func(ctx context.Context, nodeType string, msg *cmsg.Message) (cfacade.IMember, error)

func randRoute(nodeType string) (cfacade.IMember, error) {
	s := rand.NewSource(time.Now().Unix())
	rnd := rand.New(s)

	memberList := cdiscovery.ListByType(nodeType)
	if len(memberList) < 1 {
		return nil, cerr.DiscoveryMemberListIsEmpty
	}

	if len(memberList) == 1 {
		return memberList[0], nil
	}

	server := memberList[rnd.Intn(len(memberList))]
	return server, nil
}

func AddRoute(nodeType string, routingFunction RoutingFunc) {
	if _, ok := routesMap[nodeType]; ok {
		clog.Warnf("overriding the route to [nodeType = %s]", nodeType)
	}

	routesMap[nodeType] = routingFunction
}

func Route(ctx context.Context, nodeType string, msg *cmsg.Message) (cfacade.IMember, error) {
	routeFunc, ok := routesMap[nodeType]
	if !ok {
		return randRoute(nodeType)
	}

	return routeFunc(ctx, nodeType, msg)
}
