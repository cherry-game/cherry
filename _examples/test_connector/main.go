package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/gin"
	"github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/session"
)

func main() {
	app := cherry.NewApp("../test1_handler/profile/", "local", "web-1")

	httpServer := cherryGin.New("127.0.0.1:80", cherryGin.RecoveryWithZap(true))
	httpServer.StaticFS("/", "./web/")

	app.OnStartup(
		cherrySession.NewComponent(),
		cherryHandler.NewComponent(),
		httpServer,
		cherryConnector.NewWebsocketComponent("127.0.0.1:34590"),
	)

	app.OnShutdown()
}
