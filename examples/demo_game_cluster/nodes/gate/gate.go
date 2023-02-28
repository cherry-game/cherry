package gate

import (
	"github.com/cherry-game/cherry"
	cherryGops "github.com/cherry-game/cherry/components/gops"
	checkCenter "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/component/check_center"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/data"
	cconnector "github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

// Run 运行gate节点
// gate 主要用于对外提供网络连接、管理用户连接、消息转发
func Run(profileFilePath, nodeId string) {
	// 创建一个cherry实例
	app := cherry.Configure(
		profileFilePath,
		nodeId,
		true,
		cherry.Cluster,
	)

	// 设置actor组件调用函数
	app.SetActorInvoke(pomelo.LocalInvokeFunc, pomelo.RemoteInvokeFunc)

	// 使用pomelo网络数据包解析器
	agentActor := pomelo.NewActor("user")
	//创建一个tcp监听，用于client/robot压测机器人连接网关tcp
	agentActor.AddConnector(cconnector.NewTCP(":10011"))
	//再创建一个websocket监听，用于h5客户端建立连接
	agentActor.AddConnector(cconnector.NewWS(app.Address()))
	//当有新连接创建Agent时，启动一个自定义(ActorAgent)的子actor
	agentActor.SetOnNewAgent(func(newAgent *pomelo.Agent) {
		childActor := &ActorAgent{}
		newAgent.AddOnClose(childActor.OnSessionClose)
		// actorID == sid
		agentActor.Child().Create(newAgent.SID(), childActor)
	})

	// 设置数据路由函数
	agentActor.SetOnDataRoute(onDataRoute)

	app.SetNetParser(agentActor)

	app.Register(cherryGops.New())
	// 注册检则中心服组件，用于检则中心服是否先启动
	app.Register(checkCenter.New())
	// 注册数据配表组件，具体详见data-config的使用方法和参数配置
	app.Register(data.New())

	//启动cherry引擎，设置为前端类型的节点，并且以集群方式运行
	app.Startup()
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
