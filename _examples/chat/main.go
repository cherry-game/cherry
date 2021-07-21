package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/gin"
	cherryORM "github.com/cherry-game/cherry/component/orm"
	"github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
)

func main() {
	app := cherry.NewApp("../config/", "dev", "game-1")
	app.SetSerializer(cherrySerializer.NewJSON())

	httpComp := cherryGin.New("127.0.0.1:80", cherryGin.RecoveryWithZap(true))
	httpComp.StaticFS("/", "./web/")

	sessionComp := cherrySession.NewComponent()

	connectorComp := cherryConnector.NewWSComponent(app.Address())
	//connectorComp := cherryConnector.NewTCPComponent(app.Address())

	app.Startup(
		cherryORM.NewComponent(),
		sessionComp,
		handlerComponent(),
		httpComp,
		connectorComp,
	)

}

func handlerComponent() *cherryHandler.Component {
	component := cherryHandler.NewComponent(
		cherryHandler.WithPrintRouteLog(true),
	)

	group1 := cherryHandler.NewGroup(10, 256)
	group1.AddHandlers(&userHandler{})
	component.Register(group1)

	group2 := cherryHandler.NewGroup(10, 256)
	group2.AddHandlers(&roomHandler{})
	component.Register(group2)

	return component
}
