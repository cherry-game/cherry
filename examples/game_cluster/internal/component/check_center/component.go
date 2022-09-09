package checkCenter

import (
	"github.com/cherry-game/cherry"
	rpcCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/rpc/center"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"time"
)

// Component 启动时,检查center节点是否存活
type Component struct {
	cherryFacade.Component
}

func RegisterComponent() {
	cherry.RegisterComponent(&Component{})
}

func (c *Component) Name() string {
	return "run_check_component"
}

func (c *Component) OnAfterInit() {
	for {
		if rpcCenter.Ping() {
			break
		}
		time.Sleep(2 * time.Second)
		cherryLogger.Warn("center node connect fail. retrying in 2 seconds.")
	}
}
