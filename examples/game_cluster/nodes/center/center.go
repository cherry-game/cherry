package center

import (
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center/db"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center/module/account"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center/module/ops"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
)

func Run(path, name, nodeId string) {
	cherry.Configure(path, name, nodeId)

	cherryCron.RegisterComponent()
	data.RegisterComponent()
	db.RegisterComponent()

	registerHandler()

	cherry.Run(false, cherry.Cluster)
}

func registerHandler() {
	centerGroup := cherryHandler.NewGroup(60, 256)
	centerGroup.AddHandlers(
		&account.Handler{},
		&ops.Handler{},
	)

	cherry.RegisterHandlerGroup(centerGroup)
}
