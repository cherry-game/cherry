package cherryCluster

import (
	"github.com/cherry-game/cherry/facade"
)

type NodeDiscovery interface {
	All() NodeMap
	GetType(nodeId string) (nodeType string, err error)
	Get(nodeId string) cherryFacade.INode
	Sync()
	AddListener(listener NodeListener)
}

type NodeListener interface {
	Add(node cherryFacade.INode)
	Remove(node cherryFacade.INode)
}

// AddServers
// RemoveServers
// ReplaceServers
