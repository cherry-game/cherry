package gate

import (
	"context"
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	cherryError "github.com/cherry-game/cherry/error"
	checkCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/component/check_center"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	sessionKey "github.com/cherry-game/cherry/examples/game_cluster/internal/session_key"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryDiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySession "github.com/cherry-game/cherry/net/session"
)

// Run gate不直接连数据库，需要数据可以通center/game节点获取
func Run(path, name, node string) {
	gateApp := cherry.Configure(path, name, node)

	cherryCron.RegisterComponent()
	checkCenter.RegisterComponent()
	data.RegisterComponent()

	registerHandlers()
	//rateLimiter(gateApp)
	createConnector(gateApp)

	cherry.AddNodeRouter("game", gameNodeRoute)

	cherry.Run(true, cherry.Cluster)
}

func createConnector(app cherryFacade.IApplication) {
	connector := cherryConnector.NewTCP(app.Address())
	cherry.RegisterConnector(connector)
}

//// rateLimiter 限速过滤
//func rateLimiter(app cherryFacade.IApplication) {
//	enable := app.Settings().GetBool("rate_limit")
//	if enable {
//		cherrySession.AddOnDataListener(rate.SessionIPLimiter(
//			3*time.Second,
//			100,
//			code.NodeRateLimiter,
//			false,
//			180,
//		))
//	}
//}

func registerHandlers() {
	userGroup := cherryHandler.NewGroup(512, 512)
	userGroup.AddHandlers(&UserHandler{})

	cherry.RegisterHandlerGroup(userGroup)
}

// gameNodeRoute 实现网关路由消息到游戏节点的逻辑
func gameNodeRoute(ctx context.Context, route *cherryMessage.Route, session *cherrySession.Session) (cherryFacade.IMember, error) {
	if session == nil || session.IsBind() == false {
		return nil, cherryError.Error("session not bind,message forwarding is not allowed.")
	}

	// 根据session绑定的游戏id进行节点消息路由
	serverId := session.GetString(sessionKey.ServerID)
	if member, found := cherryDiscovery.GetMember(serverId); found {
		return member, nil
	}

	return nil, cherryError.DiscoveryMemberListIsEmpty
}
