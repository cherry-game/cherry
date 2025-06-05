package cherryNatsCluster

import (
	"fmt"
)

const (
	remoteSubjectFormat = "cherry.%s.remote.%s.%s" // nodeType.nodeID
	localSubjectFormat  = "cherry.%s.local.%s.%s"  // nodeType.nodeID
)

// getLocalSubject local message nats chan
func getLocalSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(localSubjectFormat, prefix, nodeType, nodeID)
}

// getRemoteSubject remote message nats chan
func getRemoteSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(remoteSubjectFormat, prefix, nodeType, nodeID)
}
