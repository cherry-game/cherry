package cherrySnowflake

import (
	ccrypto "github.com/cherry-game/cherry/extend/crypto"
	clog "github.com/cherry-game/cherry/logger"
)

var (
	defaultNode *Node
)

func InitDefaultNode(str string) {
	var (
		crc32Value = int64(ccrypto.CRC32(str))
		nodeValue  = crc32Value % nodeMax
	)

	SetDefaultNode(nodeValue)
}

func SetDefaultNode(nodeValue int64) {
	if defaultNode != nil {
		clog.Warn("default snowflake node is created.")
		return
	}

	var err error
	defaultNode, err = NewNode(nodeValue)
	if err != nil {
		clog.Warn(err)
		clog.Warnf("create default snowflake node fail. nodeValue = %d", nodeValue)
	}

	clog.Infof("[snowflake] nodeValue = %d, nodeMax = %d", nodeValue, nodeMax)
}

func Next() ID {
	return defaultNode.Generate()
}

func NextID() int64 {
	return defaultNode.Generate().Int64()
}
