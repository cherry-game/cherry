package cherrySnowflake

import clog "github.com/cherry-game/cherry/logger"

var (
	defaultNode *Node
)

func SetDefaultNode(nodeId int64) {
	if defaultNode != nil {
		clog.Warn("default snowflake node is created.")
		return
	}

	var err error
	defaultNode, err = NewNode(nodeId)
	if err != nil {
		clog.Warn(err)
		clog.Warnf("create default snowflake node fail. nodeId=%d", nodeId)
	}
}

func Next() ID {
	return defaultNode.Generate()
}

func NextId() int64 {
	return defaultNode.Generate().Int64()
}
