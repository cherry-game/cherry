package cherryProto

import (
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

func (x *ClusterPacket) Recycle() {
	x.BuildTime = 0
	x.SourcePath = ""
	x.TargetPath = ""
	x.FuncName = ""
	x.ArgBytes = nil
	x.Session = nil
	clusterPacketPool.Put(x)
}
