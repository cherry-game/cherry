package master

import (
	"github.com/cherry-game/cherry"
)

func Run(profileFilePath, nodeId string) {
	app := cherry.Configure(profileFilePath, nodeId, false, cherry.Cluster)
	app.Startup()
}
