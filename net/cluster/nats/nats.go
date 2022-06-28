package cherryNats

import (
	clog "github.com/cherry-game/cherry/logger"
	cprofile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"time"
)

var (
	natsConnect *NatsConnect
)

func Init() {
	natsConfig := cprofile.GetConfig("cluster").GetConfig("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	natsConnect = New()
	natsConnect.LoadConfig(natsConfig)
	natsConnect.Connect()

	clog.Infof("nats execute Init()")
}

func App() *NatsConnect {
	return natsConnect
}

func Publish(subj string, data []byte) error {
	return natsConnect.Publish(subj, data)
}

func Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return natsConnect.Request(subj, data, timeout)
}

func ChanSubscribe(subj string, ch chan *nats.Msg) (*nats.Subscription, error) {
	return natsConnect.ChanSubscribe(subj, ch)
}

func Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	return natsConnect.Subscribe(subj, cb)
}

func Close() {
	natsConnect.Close()
	clog.Infof("nats execute Close()")
}
