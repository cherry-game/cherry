package simple

import (
	"encoding/binary"
	"time"
)

// Remote function name constants.
const (
	ResponseFuncName = "response"
)

// Package-level configuration shared by all simple-protocol agents.
var (
	heartbeatTime                  = time.Second * 60 // heartbeat interval
	writeBacklog                   = 64               // write and pending channel buffer size
	endian        binary.ByteOrder = binary.BigEndian // byte order for encoding/decoding message headers
)

// SetHeartbeatTime sets the heartbeat interval for agent connections.
// Values less than 1 second are ignored.
func SetHeartbeatTime(t time.Duration) {
	if t.Seconds() > 1 {
		heartbeatTime = t
	}
}

// SetWriteBacklog sets the size of the write and pending channel buffers.
// Values less than or equal to 0 are ignored.
func SetWriteBacklog(backlog int) {
	if backlog > 0 {
		writeBacklog = backlog
	}
}

// SetEndian sets the byte order used for encoding/decoding message headers.
func SetEndian(e binary.ByteOrder) {
	if e != nil {
		endian = e
	}
}
