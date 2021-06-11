package main

import (
	"github.com/cherry-game/cherry"
	cherryGin "github.com/cherry-game/cherry/component/gin"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"net/http"
)

func main() {
	app := cherry.NewApp("../test1_handler/profile/", "local", "web-1")

	defer app.OnShutdown()

	httpServer := cherryGin.New("127.0.0.1:80", cherryGin.RecoveryWithZap(true))
	httpServer.GetEngine().StaticFS("/", http.Dir("F:\\bd\\cherry\\_examples\\test_connector\\web\\"))

	app.OnStartup(
		cherrySession.NewComponent(),
		cherryHandler.NewComponent(),
		httpServer,
		cherryConnector.NewWebsocketComponent("127.0.0.1:34590"),
	)

}
