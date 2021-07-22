package main

import (
	"github.com/cherry-game/cherry"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
)

func main() {
	masterApp := cherry.NewApp("../config/", "test", "master-1")

	masterApp.Startup(
		cherryCluster.NewComponent(),
	)
}
