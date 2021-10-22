package cherrySnowflake

import cherryLogger "github.com/cherry-game/cherry/logger"

var (
	defaultNode *Node
)

func SetDefaultNode(nodeId int64) {
	if defaultNode != nil {
		cherryLogger.Warn("default snowflake node is created.")
		return
	}

	var err error
	defaultNode, err = NewNode(nodeId)
	if err != nil {
		cherryLogger.Warn(err)
		cherryLogger.Warnf("create default snowflake node fail. nodeId=%d", nodeId)
	}
}

func Next() ID {
	return defaultNode.Generate()
}

func NextId() int64 {
	return defaultNode.Generate().Int64()
}
