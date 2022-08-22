package main

import (
	"github.com/cherry-game/cherry"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
)

func main() {
	gameApp := cherry.NewApp("../../config/", "discovery", "master-1")

	gameApp.Startup(
		cherryHandler.NewComponent(),
		cherryCluster.NewComponent(),
	)
}
