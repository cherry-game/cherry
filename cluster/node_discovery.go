package cherryCluster

import (
	"github.com/phantacix/cherry/interfaces"
)

type NodeDiscovery interface {
	All() NodeMap
	GetType(nodeId string) (nodeType string, err error)
	Get(nodeId string) cherryInterfaces.INode
	Sync()
	AddListener(listener NodeListener)
}

type NodeListener interface {
	Add(node cherryInterfaces.INode)
	Remove(node cherryInterfaces.INode)
}
