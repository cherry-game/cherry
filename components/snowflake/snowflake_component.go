package cherrySnowflake

import (
	cherryConst "github.com/cherry-game/cherry/const"
	cherryInterfaces "github.com/cherry-game/cherry/interfaces"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/utils/snowflake"
)

type SnowFlakeComponent struct {
	cherryInterfaces.BaseComponent
	snowflake *snowflake.Node
}

func (s *SnowFlakeComponent) Name() string {
	return cherryConst.SnowflakeComponent
}

func New(node int64) *SnowFlakeComponent {
	c := &SnowFlakeComponent{}

	var err error
	c.snowflake, err = snowflake.NewNode(node)
	if err != nil {
		cherryLogger.Warnf("create snowflake component fail. node=%d", node)
		panic(err)
	}

	return c
}
