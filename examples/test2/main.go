package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/components/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {

	testApp := cherry.DefaultApp()
	defer testApp.Shutdown()

	httpServer := cherryGin.New("http_server_1")
	httpServer.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "this is index...")
	})

	httpServer.GET("/panic", func(c *gin.Context) {
		c.String(http.StatusOK, "test panic")
		panic("test panic!")
	})

	httpServer.Run(testApp.ThisNode().Address())

	testApp.Startup(httpServer)
}
