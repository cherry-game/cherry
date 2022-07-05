package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/gin"
	cconnector "github.com/cherry-game/cherry/net/connector"
	chandler "github.com/cherry-game/cherry/net/handler"
	cserializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	app := cherry.Configure("../config/", "chat", "game-1")

	cherry.SetSerializer(cserializer.NewJSON())

	httpComp := cherryGin.New("web", "127.0.0.1:80")
	httpComp.Use(cherryGin.RecoveryWithZap(true))
	httpComp.Static("/", "./web/")
	cherry.RegisterComponent(httpComp)

	cherry.RegisterConnector(cconnector.NewWS(app.Address()))

	handlerComponent()

	cherry.Run(true, cherry.Standalone)
}

func handlerComponent() {
	group1 := chandler.NewGroup(1, 256)
	group1.AddHandlers(&userHandler{})
	group1.AddHandlers(&roomHandler{})

	cherry.RegisterHandlerGroup(group1)
}
