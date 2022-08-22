package main

import (
	"github.com/cherry-game/cherry"
	cherryGin "github.com/cherry-game/cherry/gin"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	webApp := cherry.Configure("../../config/", "sample1", "web-1")

	cherry.SetSerializer(cherrySerializer.NewJSON())

	httpComp := cherryGin.New("web", webApp.Address())
	httpComp.Use(cherryGin.RecoveryWithZap(true))
	httpComp.Static("/", "./static/")
	cherry.RegisterComponent(httpComp)

	cherry.Run(false, cherry.Standalone)
}
