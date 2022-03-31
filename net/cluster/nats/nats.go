package cherryNats

import cherryProfile "github.com/cherry-game/cherry/profile"

var (
	thisNats = NewNats()
)

func Init() {
	thisNats.loadConfig(cherryProfile.Get("cluster").Get("nats"))
	thisNats.Connect()
}

func Conn() *NatsConnect {
	return thisNats
}
