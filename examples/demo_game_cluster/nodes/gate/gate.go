package gate

import (
	"encoding/binary"
	"github.com/cherry-game/cherry"
	cherryGops "github.com/cherry-game/cherry/components/gops"
	checkCenter "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/component/check_center"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/data"
	cfacade "github.com/cherry-game/cherry/facade"
	cconnector "github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	"github.com/cherry-game/cherry/net/parser/simple"
	"time"
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

	// 设置网络数据包解析器
	netParser := buildPomeloParser(app)
	//netParser := buildSimpleParser(app)
	app.SetNetParser(netParser)

	app.Register(cherryGops.New())
	// 注册检则中心服组件，用于检则中心服是否先启动
	app.Register(checkCenter.New())
	// 注册数据配表组件，具体详见data-config的使用方法和参数配置
	app.Register(data.New())

	//启动cherry引擎
	app.Startup()
}

func buildPomeloParser(app *cherry.AppBuilder) cfacade.INetParser {
	// 使用pomelo网络数据包解析器
	agentActor := pomelo.NewActor("user")
	//创建一个tcp监听，用于client/robot压测机器人连接网关tcp
	agentActor.AddConnector(cconnector.NewTCP(":10011"))
	//再创建一个websocket监听，用于h5客户端建立连接
	agentActor.AddConnector(cconnector.NewWS(app.Address()))
	//当有新连接创建Agent时，启动一个自定义(ActorAgent)的子actor
	agentActor.SetOnNewAgent(func(newAgent *pomelo.Agent) {
		childActor := &ActorAgent{}
		newAgent.AddOnClose(childActor.onSessionClose)
		agentActor.Child().Create(newAgent.SID(), childActor) // actorID == sid
	})

	// 设置数据路由函数
	agentActor.SetOnDataRoute(onPomeloDataRoute)

	return agentActor
}

// 构建简单的网络数据包解析器
func buildSimpleParser(app *cherry.AppBuilder) cfacade.INetParser {
	agentActor := simple.NewActor("user")
	agentActor.AddConnector(cconnector.NewTCP(":10011"))
	agentActor.AddConnector(cconnector.NewWS(app.Address()))

	agentActor.SetOnNewAgent(func(newAgent *simple.Agent) {
		childActor := &ActorAgent{}
		//newAgent.AddOnClose(childActor.onSessionClose)
		agentActor.Child().Create(newAgent.SID(), childActor)
	})

	// 设置大头&小头
	agentActor.SetEndian(binary.LittleEndian)
	// 设置心跳时间
	agentActor.SetHeartbeatTime(60 * time.Second)
	// 设置积压消息数量
	agentActor.SetWriteBacklog(64)

	// 设置数据路由函数
	//agentActor.SetOnDataRoute(onSimpleDataRoute)

	// 设置消息节点路由(建议配合data-config组件进行使用)
	// mid = 1 的消息路由到  gate节点.user的Actor.login函数上
	agentActor.AddNodeRoute(1, &simple.NodeRoute{
		NodeType: "gate",
		ActorID:  "user",
		FuncName: "login",
	})

	return agentActor
}
