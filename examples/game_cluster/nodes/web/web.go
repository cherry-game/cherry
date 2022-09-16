package web

import (
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	cherryGin "github.com/cherry-game/cherry/components/gin"
	checkCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/component/check_center"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/web/controller"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/web/sdk"
	cherryFile "github.com/cherry-game/cherry/extend/file"
	"github.com/gin-gonic/gin"
)

func Run(path, name, node string) {
	// 配置cherry引擎,加载profile配置文件
	webApp := cherry.Configure(path, name, node)

	// 注册调度组件
	cherryCron.RegisterComponent()

	// 注册检查中心服是否启动组件
	checkCenter.RegisterComponent()

	// 注册数据配表组件
	data.RegisterComponent()

	// 加载http server组件
	httpServerComponent(webApp.Address())

	// 加载sdk逻辑
	sdk.Init()

	// 启动cherry引擎
	cherry.Run(false, cherry.Cluster)
}

func httpServerComponent(addr string) {
	gin.SetMode(gin.DebugMode)

	// new http server
	httpServer := cherryGin.NewHttp("http_server", addr)
	httpServer.Use(cherryGin.Cors())

	// http server使用gin组件搭建，这里增加一个RecoveryWithZap中间件
	httpServer.Use(cherryGin.RecoveryWithZap(true))

	// 映射h5客户端静态文件到static目录，例如：http://127.0.0.1/static/protocol.js
	httpServer.Static("/static", "./static/")

	// 加载./view目录的html模板文件(详情查看gin文档)
	viewFiles := cherryFile.WalkFiles("./view/", ".html")
	if len(viewFiles) < 1 {
		panic("view files not found.")
	}
	httpServer.LoadHTMLFiles(viewFiles...)

	//注册 controller
	httpServer.Register(new(controller.Controller))

	// 添加httpServer组件
	cherry.RegisterComponent(httpServer)
}
