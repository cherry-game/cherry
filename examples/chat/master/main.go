package main

import (
	"github.com/cherry-game/cherry"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	cherry.Configure("./examples/config/", "chat", "chat-master")
	cherry.SetSerializer(cserializer.NewJSON())
	cherry.Run(false, cherry.Cluster)
}
