package main

import (
	"github.com/cherry-game/cherry"
	cherryGin "github.com/cherry-game/cherry/components/gin"
)

func main() {
	app := cherry.NewApp(
		"./examples/config/profile-local.json",
		"web-1",
		false,
		cherry.Standalone,
	)

	httpServer := cherryGin.NewHttp("web_1", app.Address())
	httpServer.Use(cherryGin.Cors(), cherryGin.MaxConnect(2))
	httpServer.Register(new(Test1Controller))

	app.Register(httpServer)
	app.Startup()
}
