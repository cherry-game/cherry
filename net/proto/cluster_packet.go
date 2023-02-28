package cherryProto

import (
	"fmt"
	"sync"
)

var (
	clusterPacketPool = &sync.Pool{
		New: func() interface{} {
			return new(ClusterPacket)
		},
	}
)

func GetClusterPacket() *ClusterPacket {
	return clusterPacketPool.Get().(*ClusterPacket)
}

func (m *ClusterPacket) Recycle() {
	m.Reset()
	clusterPacketPool.Put(m)
}

func (m *ClusterPacket) PrintLog() string {
	return fmt.Sprintf("source = %s, target = %s, funcName = %s, bytesLen = %d, session = %v",
		m.SourcePath,
		m.TargetPath,
		m.FuncName,
		len(m.ArgBytes),
		m.Session,
	)
}
