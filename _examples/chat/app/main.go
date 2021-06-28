package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/chat"
	"github.com/cherry-game/cherry/component/gin"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	"github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/handler"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
	"time"
)

func main() {
	app := cherry.NewApp("../profile/", "local", "gate-1")

	httpServer := cherryGin.New("127.0.0.1:80", cherryGin.RecoveryWithZap(true))
	httpServer.StaticFS("/", "../web/")

	handlerComponent := createHandler()

	app.Startup(
		cherrySession.NewComponent(),
		handlerComponent,
		httpServer,
		createWebsocket(),
	)
}

func createHandler() *cherryHandler.Component {
	component := cherryHandler.NewComponent()

	group1 := cherryHandler.NewGroup(30, 256)
	group1.AddHandlers(&chat.UserHandler{})
	component.Register(group1)

	group2 := cherryHandler.NewGroup(30, 256)
	group2.AddHandlers(&chat.RoomHandler{})
	component.Register(group2)

	// add room handler

	return component
}

func createWebsocket() *cherryConnector.Component {
	connector := cherryConnector.NewWS("127.0.0.1:34590")

	component := cherryConnector.NewComponentWithOpt(cherryAgent.Options{
		Heartbeat:        60 * time.Second,
		DataCompression:  false,
		PacketDecoder:    cherryPacket.NewPomeloDecoder(),
		PacketEncoder:    cherryPacket.NewPomeloEncoder(),
		Serializer:       cherrySerializer.NewJSON(),
		PacketListener:   make(map[cherryPacket.Type]cherryAgent.PacketListener),
		RPCHandler:       nil,
		OnCreateListener: make([]cherryAgent.SessionListener, 0),
		OnCloseListener:  make([]cherryAgent.SessionListener, 0),
	}, connector)

	return component
}
