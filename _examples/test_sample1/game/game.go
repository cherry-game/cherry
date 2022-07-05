package main

import (
	"github.com/cherry-game/cherry"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	cherry.Configure("../../config/", "sample1", "1000")
	cherry.SetSerializer(cherrySerializer.NewJSON())

	cherry.RegisterHandler(&ActorHandler{})

	cherry.Run(false, cherry.Cluster)
}
