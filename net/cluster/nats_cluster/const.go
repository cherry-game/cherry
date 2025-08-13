package cherryNatsCluster

import (
	"fmt"
)

const (
	remoteSubjectFormat = "cherry.%s.remote.%s.%s" // cherry.{prefix}.remote.{nodeType}.{nodeID}
	localSubjectFormat  = "cherry.%s.local.%s.%s"  // cherry.{prefix}.local.{nodeType}.{nodeID}
	replySubjectFormat  = "cherry.%s.reply.%s.%s"  // cherry.{prefix}.reply.{nodeType}.{nodeID}
)

// getLocalSubject local message nats chan
func getLocalSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(localSubjectFormat, prefix, nodeType, nodeID)
}

// getRemoteSubject remote message nats chan
func getRemoteSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(remoteSubjectFormat, prefix, nodeType, nodeID)
}

func getReplySubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(replySubjectFormat, prefix, nodeType, nodeID)
}
