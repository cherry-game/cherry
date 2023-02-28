package main

import (
	"github.com/cherry-game/cherry"
	cherryCode "github.com/cherry-game/cherry/code"
	cherryGin "github.com/cherry-game/cherry/components/gin"
	cconnector "github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	pmessage "github.com/cherry-game/cherry/net/parser/pomelo/message"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

// 启动main函数运行聊天室程序
func main() {
	// 配置cherry引擎
	// path       profile配置的文件路径
	// nodeId     节点id,每个节点都有唯一的节点id，并且他们归属于某一个节点类型(nodeType)
	// isFrontend 节点为前端类型，则可使用connector连接器组件
	// nodeMode   节点模式为集群
	path := "./examples/config/profile-chat.json"
	nodeId := "room-1"
	app := cherry.Configure(
		path,
		nodeId,
		true,
		cherry.Cluster,
	)

	// 设置json做为数据序列化的方式，系统默认:protobuf
	// json			cserializer.NewJSON()
	// protobuf		cserializer.NewProtobuf()
	app.SetSerializer(cserializer.NewJSON())

	// 设置actor组件调用函数
	app.SetActorInvoke(pomelo.LocalInvokeFunc, pomelo.RemoteInvokeFunc)

	// 创建pomelo网络数据包解析器，它同时也是一个actor
	agentActor := pomelo.NewActor("user")
	// 添加websocket连接器, 根据业务需要可添加多类型的connector
	agentActor.AddConnector(cconnector.NewWS(":34590"))
	// 创建Agent时，启动一个自定义(actorUser)的子actor
	agentActor.SetOnNewAgent(func(newAgent *pomelo.Agent) {
		newUser := newActorUser(newAgent)
		agentActor.Child().Create(newAgent.SID(), &newUser)
	})

	// 设置数据路由函数
	agentActor.SetOnDataRoute(onDataRoute)

	// 当建立新的客户端连接时，默认使用系统自带的处理方式
	//agentActor.SetOnConnect()
	// 设置网络包解析器
	app.SetNetParser(agentActor)

	//添加actor
	app.AddActors(
		&actorRoom{},
		&actorWrite{},
	)

	// 启动http server
	httpServerComponent(app)

	// 运行cherry引擎
	app.Startup()
}

func onDataRoute(agent *pomelo.Agent, route *pmessage.Route, msg *pmessage.Message) {
	// 登录消息可继续进行
	if msg.Route == "room.user.login" {
		pomelo.DefaultDataRoute(agent, route, msg)
		return
	}

	// session未绑定uid，拦截消息
	if !agent.Session().IsBind() {
		agent.ResponseCode(agent.Session(), cherryCode.SessionUIDNotBind)
		return
	}

	pomelo.DefaultDataRoute(agent, route, msg)
}

// 为了省事，构造一个http server用于部署我们的客户端h5静态文件
func httpServerComponent(app *cherry.AppBuilder) {
	// 启动后访问 http://127.0.0.1/ 即可
	httpComp := cherryGin.New("web", "127.0.0.1:80")
	// http server使用gin组件搭建，这里增加一个RecoveryWithZap中间件
	httpComp.Use(cherryGin.RecoveryWithZap(true))

	// 直接映射h5客户端静态文件到根目录
	httpComp.Static("/", "../static/")
	// 把http server组件注册到cherry引擎中
	app.Register(httpComp)
}
