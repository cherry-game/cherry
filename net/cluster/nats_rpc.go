package cherryCluster

import (
	"fmt"
)

const (
	nodeLocalSubjectFormat  = "cherry.node.local.%s.%s"
	nodeRemoteSubjectFormat = "cherry.node.remote.%s.%s"
)

// GetLocalNodeSubject local packet nats chan
func GetLocalNodeSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(nodeLocalSubjectFormat, nodeType, nodeId)
}

// GetRemoteNodeSubject remote packet nats chan
func GetRemoteNodeSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(nodeRemoteSubjectFormat, nodeType, nodeId)
}
