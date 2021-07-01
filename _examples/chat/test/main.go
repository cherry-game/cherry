package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/chat/room"
	"github.com/cherry-game/cherry/_examples/chat/user"
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/agent"
	"github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	cherryPacket "github.com/cherry-game/cherry/net/packet"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
	"time"
)

func main() {

	app := cherry.NewApp("../profile/", "local", "gate-1")
	app.SetSerializer(cherrySerializer.NewJSON())

	handlerComponent := createHandler()

	go func() {
		time.Sleep(10 * time.Second)

		//agent := cherryAgent.NewAgent(app, cherryAgent.Options{}, nil)

		session := &cherrySession.Session{}

		msg := &cherryMessage.Message{
			Route: "gate.userHandler.testLogin",
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

			handlerComponent.PostMessage(session, msg)
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
	group1.AddHandlers(&user.Handler{})
	component.Register(group1)

	group2 := cherryHandler.NewGroup(1, 256)
	group2.AddHandlers(&room.Handler{})
	component.Register(group2)

	// add room handler

	return component
}

func createWebsocket() *cherryConnector.Component {
	connector := cherryConnector.NewWS("127.0.0.1:34590")

	component := cherryConnector.NewComponentWithOpt(cherryAgent.Options{
		Heartbeat:       60 * time.Second,
		DataCompression: false,
		PacketListener:  make(map[cherryPacket.Type]cherryAgent.PacketListener),
		RPCHandler:      nil,
	}, connector)

	return component
}
