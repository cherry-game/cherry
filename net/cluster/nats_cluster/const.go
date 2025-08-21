package cherryNatsCluster

import (
	"fmt"
)

const (
	Asterisk = "*"
)

const (
	LocalType  = "local"
	RemoteType = "remote"
)

const (
	remoteSubjectFormat = "cherry.%s.remote.%s.%s" // cherry.{prefix}.remote.{nodeType}.{nodeID}
	localSubjectFormat  = "cherry.%s.local.%s.%s"  // cherry.{prefix}.local.{nodeType}.{nodeID}
	replySubjectFormat  = "cherry.%s.reply.%s.%s"  // cherry.{prefix}.reply.{nodeType}.{nodeID}
)

// GetLocalSubject local message nats chan
func GetLocalSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(localSubjectFormat, prefix, nodeType, nodeID)
}

// GetRemoteSubject remote message nats chan
func GetRemoteSubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(remoteSubjectFormat, prefix, nodeType, nodeID)
}

func GetReplySubject(prefix, nodeType, nodeID string) string {
	return fmt.Sprintf(replySubjectFormat, prefix, nodeType, nodeID)
}
