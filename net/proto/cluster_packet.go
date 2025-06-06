package cherryProto

import (
	"fmt"
	"sync"

	ctime "github.com/cherry-game/cherry/extend/time"
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
	pkg.BuildTime = ctime.Now().ToMillisecond()
	return pkg
}

func BuildClusterPacket(source, target, funcName string) *ClusterPacket {
	clusterPacket := GetClusterPacket()
	clusterPacket.SourcePath = source
	clusterPacket.TargetPath = target
	clusterPacket.FuncName = funcName
	return clusterPacket
}

func (x *ClusterPacket) Recycle() {
	x.BuildTime = 0
	x.SourcePath = ""
	x.TargetPath = ""
	x.FuncName = ""
	x.ArgBytes = nil
	x.Session = nil
	clusterPacketPool.Put(x)
}

func (x *ClusterPacket) PrintLog() string {
	return fmt.Sprintf("source = %s, target = %s, funcName = %s, bytesLen = %d, session = %v",
		x.SourcePath,
		x.TargetPath,
		x.FuncName,
		len(x.ArgBytes),
		x.Session,
	)
}
