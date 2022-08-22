package main

import (
	"github.com/cherry-game/cherry"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
)

func main() {
	gameApp := cherry.NewApp("../../config/", "discovery", "game-2")

	gameApp.Startup(
		cherryHandler.NewComponent(),
		cherryCluster.NewComponent(),
	)
}
