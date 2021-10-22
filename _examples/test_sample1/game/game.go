package main

import (
	"github.com/cherry-game/cherry"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	cherry.Configure("../../config/", "sample1", "1000")
	cherry.SetSerializer(cherrySerializer.NewJSON())

	cherry.RegisterHandler(&ActorHandler{})
	cherry.SetHandlerOptions(cherryHandler.WithPrintRouteLog(true))

	cherry.Run(false, cherry.Cluster)
}
