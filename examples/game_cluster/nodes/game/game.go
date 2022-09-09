package game

import (
	"github.com/cherry-game/cherry"
	cherryCron "github.com/cherry-game/cherry/components/cron"
	checkCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/component/check_center"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/center/db"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/game/module/actor"
	"github.com/cherry-game/cherry/examples/game_cluster/nodes/game/sessions"
	cherrySnowflake "github.com/cherry-game/cherry/extend/snowflake"
	cherryUtils "github.com/cherry-game/cherry/extend/utils"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
)

func Run(path, name, nodeId string) {
	if cherryUtils.IsNumeric(nodeId) == false {
		panic("node parameter must is number.")
	}

	// set current serverId
	sessions.SetServerId(nodeId)
	// snowflake global id
	cherrySnowflake.SetDefaultNode(int64(sessions.ServerId()))
	// 配置cherry引擎
	cherry.Configure(path, name, nodeId)
	// 注册调度组件
	cherryCron.RegisterComponent()
	// 注册检测中心节点组件，确认中心节点启动后，再启动当前节点
	checkCenter.RegisterComponent()
	// 注册数据配置组件
	data.RegisterComponent()
	// 注册db组件
	db.RegisterComponent()

	registerHandlers()

	cherry.Run(false, cherry.Cluster)
}

// registerHandlers 注册所有handler
func registerHandlers() {
	// 角色handler组(本组handler都在一个协程池管理)
	actorGroup := cherryHandler.NewGroup(512, 512)
	actorGroup.AddHandlers(
		&actor.Handler{},
		//&chat.Handler{},
		//&hero.Handler{},
		//&item.Handler{},
		//&mail.Handler{},
		//&activity.Handler{},
	)

	cherry.RegisterHandlerGroup(actorGroup)
}
