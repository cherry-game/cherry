package main

import (
	"github.com/cherry-game/cherry"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	cherry.Configure("./examples/config/", "chat", "log-1")
	cherry.SetSerializer(cserializer.NewJSON())
	cherry.RegisterHandler(&Handler{})
	cherry.Run(false, cherry.Cluster)
}
