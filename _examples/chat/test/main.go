package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/chat"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherryRoute "github.com/cherry-game/cherry/net/route"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"time"
)

func main() {
	app := cherry.NewApp("../profile/", "local", "gate-1")

	handlerComponent := createHandler()

	go func() {
		time.Sleep(10 * time.Second)

		route := cherryRoute.NewByName("gate.BindService.Login")
		agent := &cherryAgent.Agent{
			Options: cherryAgent.Options{
				Serializer: cherrySerializer.NewJSON(),
			},
			Session: &cherrySession.Session{},
		}

		msg := &cherryMessage.Message{
			Data: []byte{
				123, 34, 110, 105, 99, 107, 110, 97, 109, 101,
				34, 58, 34, 103, 117, 101, 115, 116, 49, 54,
				50, 52, 54, 50, 48, 57, 57, 54, 55, 49, 50, 34, 125,
			},
		}

		i := 0

		for {
			if app.Running() == false {
				return
			}

			handlerComponent.PostMessage(agent, route, msg)
			//time.Sleep(time.Microsecond * 1)

			i++

			if i%100000 == 1 {
				cherryLogger.Infof("count num = %d, time = %d", i, cherryTime.Now().ToMillisecond())
			}
		}

	}()

	app.Startup(
		cherrySession.NewComponent(),
		handlerComponent,
		createWebsocket(),
	)
}

func createHandler() *cherryHandler.Component {
	component := cherryHandler.NewComponent()

	group1 := cherryHandler.NewGroup(1, 256)
	group1.AddHandlers(&chat.UserHandler{})
	component.Register(group1)

	group2 := cherryHandler.NewGroup(1, 256)
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
