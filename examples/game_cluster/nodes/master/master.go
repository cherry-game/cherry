package master

import (
	"github.com/cherry-game/cherry"
)

func Run(path, name, nodeId string) {
	cherry.Configure(path, name, nodeId)
	cherry.Run(false, cherry.Cluster)
}
