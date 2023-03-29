package cherryProto

import (
	"fmt"
	"sync"
	"time"
)

var (
	clusterPacketPool = &sync.Pool{
		New: func() interface{} {
			return new(ClusterPacket)
		},
	}
)

func GetClusterPacket() *ClusterPacket {
	pkg := clusterPacketPool.Get().(*ClusterPacket)
	pkg.BuildTime = time.Now().UnixMilli()
	return pkg
}

func (m *ClusterPacket) Recycle() {
	m.BuildTime = 0
	m.SourcePath = ""
	m.TargetPath = ""
	m.FuncName = ""
	m.ArgBytes = nil
	m.Session = nil
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
