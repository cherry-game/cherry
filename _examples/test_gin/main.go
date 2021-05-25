package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/gin"
)

func main() {
	testApp := cherry.NewDefaultApp()
	defer testApp.OnShutdown()

	httpServer := cherryGin.NewHttp("http_server_1", testApp.ThisNode().Address())
	httpServer.Register(new(Test1Controller))

	testApp.OnStartup(httpServer)
}
