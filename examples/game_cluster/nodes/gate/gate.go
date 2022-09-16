package gate

import (
	"context"
	"github.com/cherry-game/cherry"
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

// Run 运行gate节点
// gate 主要用于对外提供网络连接、管理用户连接、消息转发
func Run(path, name, node string) {
	// 创建一个cherry实例
	gateApp := cherry.Configure(path, name, node)

	// 默认序列类型为protobuf,所以没有调用 cherry.SetSerializer()

	// 注册检则中心服组件，主要用于检则中心服是否先启动
	checkCenter.RegisterComponent()
	// 注册数据配表组件，具体详见data-config的使用方法和参数配置
	data.RegisterComponent()

	// 注册gate handler，客户端连接上来后，首先是帐号登陆验证token
	registerHandlers()
	//rateLimiter(gateApp)

	// 创建一个连接器监听，用于客户端连接该网关端口
	createConnector(gateApp)

	//添加game节点路由,所有转到到game节点的消息通过gameNodeRoute函数来处理
	cherry.AddNodeRouter("game", gameNodeRoute)

	//启动cherry引擎，设置为前端类型的节点，并且以集群方式运行
	cherry.Run(true, cherry.Cluster)
}

func createConnector(app cherryFacade.IApplication) {
	// TODO 网关可以同时添加多个连接器,方便各种类型协议进行通信(比如做h5、手游多端游戏产品)

	//创建一个tcp监听，用于client/robot压测机器人连接网关tcp,protobuf
	tcpConnector := cherryConnector.NewTCP(":10011")
	cherry.RegisterConnector(tcpConnector)

	//再创建一个websocket监听，用于h5客户端建立连接
	wsConnector := cherryConnector.NewWS(app.Address())
	cherry.RegisterConnector(wsConnector)
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

// gameNodeRoute 实现网关路由消息到游戏节点的函数
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
