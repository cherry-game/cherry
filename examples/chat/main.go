package main

import (
	"github.com/cherry-game/cherry"
	cherryGin "github.com/cherry-game/cherry/components/gin"
	cconnector "github.com/cherry-game/cherry/net/connector"
	chandler "github.com/cherry-game/cherry/net/handler"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

// 启动main函数运行聊天室程序
func main() {
	// 配置cherry引擎
	// @profilePath 为profile的配置路径,
	// @profileName 为profile的环境名称,这里配置值为chat则表示会读取 ../config/profile-chat.json文件
	// @nodeId 		节点id,每个节点都有一个唯一的节点id，并且他们归属于某一个节点类型(nodeType)
	app := cherry.Configure("../config/", "chat", "game-1")

	// 设置json作为客户端通信时序列化的格式,详见protocol.go定义的结构
	// 引擎自带json和protobuf，默认值启用protobuf,也可根据实际情况进行扩展
	// json			cserializer.NewJSON()
	// protobuf		cserializer.NewProtobuf()
	cherry.SetSerializer(cserializer.NewJSON())

	// 注册http server组件
	httpServerComponent()

	// 注册一个websocket的连接器组件到cherry引擎中
	// h5客户端启动后会连接该端口与服务器端进行通信,具体看 ./static/index.html的42行
	cherry.RegisterConnector(cconnector.NewWS(app.Address()))

	// 注册handler组件
	handlerComponent()

	// 运行chery引擎
	// 节点为前端类型，则可使用connector连接器组件
	// 节点模式为单机
	cherry.Run(true, cherry.Standalone)
}

// 为了省事，构造一个http server用于部署我们的客户端h5静态文件
func httpServerComponent() {
	// 启动后访问 http://127.0.0.1/ 即可
	httpComp := cherryGin.New("web", "127.0.0.1:80")
	// http server使用gin组件搭建，这里增加一个RecoveryWithZap中间件
	httpComp.Use(cherryGin.RecoveryWithZap(true))

	// 直接映射h5客户端静态文件到根目录
	httpComp.Static("/", "./static/")
	// 把http server组件注册到cherry引擎中
	cherry.RegisterComponent(httpComp)
}

// 注册服务器端的处理函数，handler用于接收客户端的请求，并进行处理
func handlerComponent() {
	// 创建一个handler处理组
	group1 := chandler.NewGroup(1, 256)
	// 加入 用户处理handler
	group1.AddHandlers(&userHandler{})
	// 加入 房间处理handler
	group1.AddHandlers(&roomHandler{})
	// 注册 handlerGroup
	cherry.RegisterHandlerGroup(group1)
}
