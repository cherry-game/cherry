package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/gin"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/serializer"
)

func main() {
	app := cherry.Configure("../config/", "chat", "game-1")

	cherry.SetSerializer(cherrySerializer.NewJSON())

	httpComp := cherryGin.New("web", "127.0.0.1:80")
	httpComp.Use(cherryGin.RecoveryWithZap(true))

	httpComp.StaticFS("/", "./web/")
	cherry.RegisterComponent(httpComp)

	cherry.RegisterConnector(cherryConnector.NewWS(app.Address()))
	//cherry.RegisterConnector(cherryConnector.NewTCP(app.Address())

	handlerComponent()

	cherry.Run(true, cherry.Standalone)
}

func handlerComponent() {
	cherry.SetHandlerOptions(
		cherryHandler.WithPrintRouteLog(true),
	)

	group1 := cherryHandler.NewGroup(10, 256)
	group1.AddHandlers(&userHandler{})

	cherry.RegisterHandlerGroup(group1)

	group2 := cherryHandler.NewGroup(10, 256)
	group2.AddHandlers(&roomHandler{})
	cherry.RegisterHandlerGroup(group2)
}
