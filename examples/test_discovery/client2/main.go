package main

import (
	"github.com/cherry-game/cherry"
	cherryCluster "github.com/cherry-game/cherry/net/cluster"
)

func main() {
	app := cherry.NewApp(
		"./examples/config/profile-discovery.json",
		"game-2",
		false,
		cherry.Cluster,
	)
	app.Register(cherryCluster.New())
	app.Startup()
}
