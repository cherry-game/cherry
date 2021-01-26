package cherrySnowflake

import (
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
)

type SnowFlakeComponent struct {
	cherryInterfaces.BaseComponent
	snowflake *Node
}

func (s *SnowFlakeComponent) Name() string {
	return cherryConst.SnowflakeComponent
}

func New(node int64) *SnowFlakeComponent {
	c := &SnowFlakeComponent{}

	var err error
	c.snowflake, err = NewNode(node)
	if err != nil {
		cherryLogger.Warnf("create snowflake component fail. node=%d", node)
		panic(err)
	}

	return c
}
