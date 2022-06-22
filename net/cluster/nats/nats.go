package cherryNats

import (
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
	"time"
)

var (
	natsConnect *NatsConnect
)

func Init(opts ...OptionFunc) {
	natsConfig := cherryProfile.Get("cluster").Get("nats")
	if natsConfig.LastError() != nil {
		panic("cluster->nats config not found.")
	}

	natsConnect = New(natsConfig, opts...)
	natsConnect.Connect()

	cherryLogger.Infof("nats execute Init()")
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
	cherryLogger.Infof("nats execute Close()")
}
