package web

import (
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	cherryGin "github.com/cherry-game/cherry/components/gin"
	checkCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/component/check_center"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/web/controller"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/web/sdk"
)

func Run(path, name, node string) {
	webApp := cherry.Configure(path, name, node)

	cherryCron.RegisterComponent()
	checkCenter.RegisterComponent()
	data.RegisterComponent()

	// new http server
	httpServer := cherryGin.NewHttp("http_server", webApp.Address())
	httpServer.Use(cherryGin.Cors())

	httpServer.Register(new(controller.Controller))
	cherry.RegisterComponent(httpServer)

	sdk.Init()

	cherry.Run(false, cherry.Cluster)
}
