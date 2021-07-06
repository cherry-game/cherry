package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/chat/room"
	"github.com/cherry-game/cherry/_examples/chat/user"
	"github.com/cherry-game/cherry/component/gin"
	"github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
)

func main() {
	app := cherry.NewApp("./profile/", "local", "gate-1")
	app.SetSerializer(cherrySerializer.NewJSON())

	httpComp := cherryGin.New("127.0.0.1:80", cherryGin.RecoveryWithZap(true))
	httpComp.StaticFS("/", "./web/")

	sessionComp := cherrySession.NewComponent()

	wsComp := cherryConnector.NewWSComponent(app.Address())

	app.Startup(
		sessionComp,
		handlerComponent(),
		httpComp,
		wsComp,
	)
}

func handlerComponent() *cherryHandler.Component {
	component := cherryHandler.NewComponent()
	component.PrintRouteLog(true)

	group1 := cherryHandler.NewGroup(10, 256)
	group1.AddHandlers(&user.Handler{})
	component.Register(group1)

	group2 := cherryHandler.NewGroup(10, 256)
	group2.AddHandlers(&room.Handler{})
	component.Register(group2)

	return component
}
