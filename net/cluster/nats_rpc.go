package cherryCluster

import (
	"fmt"
)

const (
	nodeLocalSubjectFormat  = "cherry.node.local.%s.%s"  // nodeType.nodeId
	nodeRemoteSubjectFormat = "cherry.node.remote.%s.%s" // nodeType.nodeId
	nodePushSubjectFormat   = "cherry.node.push.%s.%s"   // nodeType.nodeId
	nodeKickSubjectFormat   = "cherry.node.kick.%s.%s"   // nodeType.nodeId
)

// getLocalSubject local message nats chan
func getLocalSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(nodeLocalSubjectFormat, nodeType, nodeId)
}

// getRemoteSubject remote message nats chan
func getRemoteSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(nodeRemoteSubjectFormat, nodeType, nodeId)
}

// getPushSubject push message nats chan
func getPushSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(nodePushSubjectFormat, nodeType, nodeId)
}

// getKickSubject kick message nats chan
func getKickSubject(nodeType string, nodeId string) string {
	return fmt.Sprintf(nodeKickSubjectFormat, nodeType, nodeId)
}
