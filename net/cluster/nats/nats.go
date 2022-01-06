package cherryNats

import cherryProfile "github.com/cherry-game/cherry/profile"

var (
	thisNats = NewNats()
)

func Init() {
	cluster := cherryProfile.Config().Get("cluster")
	thisNats.loadConfig(cluster.Get("nats"))
	thisNats.Connect()
}

func Conn() *NatsConnect {
	return thisNats
}
