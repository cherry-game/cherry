package cherryNatsCluster

import (
	"fmt"
)

const (
	remoteSubjectFormat = "cherry.node.remote.%s.%s" // nodeType.nodeId
	localSubjectFormat  = "cherry.node.local.%s.%s"  // nodeType.nodeId
)

// getLocalSubject local message nats chan
func getLocalSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(localSubjectFormat, nodeType, nodeId)
}

// getRemoteSubject remote message nats chan
func getRemoteSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(remoteSubjectFormat, nodeType, nodeId)
}
