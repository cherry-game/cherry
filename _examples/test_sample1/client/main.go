package main

import (
	"github.com/cherry-game/cherry"
	cherryGin "github.com/cherry-game/cherry/component/gin"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
)

func main() {
	webApp := cherry.Configure("../../config/", "sample1", "web-1")

	cherry.SetSerializer(cherrySerializer.NewJSON())

	httpComp := cherryGin.New(webApp.Address(), cherryGin.RecoveryWithZap(true))
	httpComp.StaticFS("/", "./static/")
	cherry.RegisterComponent(httpComp)

	cherry.Run(false, cherry.Standalone)
}
