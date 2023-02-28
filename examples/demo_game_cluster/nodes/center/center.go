package center

import (
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/center/db"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/center/module/account"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/nodes/center/module/ops"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

func Run(profileFilePath, nodeId string) {
	app := cherry.Configure(
		profileFilePath,
		nodeId,
		false,
		cherry.Cluster,
	)

	// 设置actor组件调用函数
	app.SetActorInvoke(pomelo.LocalInvokeFunc, pomelo.RemoteInvokeFunc)

	app.Register(cherryCron.New())
	app.Register(data.New())
	app.Register(db.New())

	app.AddActors(
		&account.ActorAccount{},
		&ops.ActorOps{},
	)

	app.Startup()
}
