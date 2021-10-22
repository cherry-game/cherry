package main

import (
	"github.com/cherry-game/cherry"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
)

func main() {
	gameApp := cherry.NewApp("../../config/", "discovery", "game-2")

	gameApp.Startup(
		cherryCluster.NewComponent(nil),
	)
}
