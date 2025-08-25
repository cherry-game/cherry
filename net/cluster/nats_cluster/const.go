package cherryNatsCluster

import (
	"fmt"
)

const (
	localSubjectFormat      = "cherry-%s.local.%s.%s"   // cherry.{prefix}.local.{nodeType}.{nodeID}
	remoteSubjectFormat     = "cherry-%s.remote.%s.%s"  // cherry.{prefix}.remote.{nodeType}.{nodeID}
	remoteTypeSubjectFormat = "cherry-%s.remoteType.%s" // cherry.{prefix}.remoteType.{nodeType}
	replySubjectFormat      = "cherry-%s.reply.%s.%s"   // cherry.{prefix}.reply.{nodeType}.{nodeID}

)

// GetLocalSubject local message nats chan
func GetLocalSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(localSubjectFormat, prefix, nodeType, nodeID)
}

// GetRemoteSubject remote message nats chan
func GetRemoteSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(remoteSubjectFormat, prefix, nodeType, nodeID)
}

func GetRemoteTypeSubject(prefix, nodeType string) string {
	return fmt.Sprintf(remoteTypeSubjectFormat, prefix, nodeType)
}

func GetReplySubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(replySubjectFormat, prefix, nodeType, nodeID)
}
