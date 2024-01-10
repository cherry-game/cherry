package cherryNatsCluster

import (
	"fmt"
)

const (
	remoteSubjectFormat = "cherry.%s.remote.%s.%s" // nodeType.nodeId
	localSubjectFormat  = "cherry.%s.local.%s.%s"  // nodeType.nodeId
)

// getLocalSubject local message nats chan
func getLocalSubject(prefix, nodeType, nodeId string) string {
	return fmt.Sprintf(localSubjectFormat, prefix, nodeType, nodeId)
}

// getRemoteSubject remote message nats chan
func getRemoteSubject(prefix, nodeType, nodeId string) string {
	return fmt.Sprintf(remoteSubjectFormat, prefix, nodeType, nodeId)
}
