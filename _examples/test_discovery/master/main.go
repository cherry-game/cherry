package main

import (
	"github.com/cherry-game/cherry"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
)

func main() {
	gameApp := cherry.NewApp("../../config/", "discovery", "master-1")

	gameApp.Startup(
		cherryCluster.NewComponent(nil),
	)
}
