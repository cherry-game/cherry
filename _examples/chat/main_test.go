package main

import (
	"github.com/cherry-game/cherry"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	facade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryConnector "github.com/cherry-game/cherry/net/connector"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	cherrySession "github.com/cherry-game/cherry/net/session"
	"testing"
	"time"
)

func TestPostMessage(t *testing.T) {
	app := cherry.NewApp("../config/", "local", "game-1")
	app.SetSerializer(cherrySerializer.NewJSON())

	handlerComponent := createHandler()

	go func() {
		time.Sleep(5 * time.Second)

		session := &cherrySession.Session{}

		msg := &cherryMessage.Message{
			Route: "game.userHandler.testLogin",
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
	group1.AddHandlers(&userHandler{})
	component.Register(group1)

	group2 := cherryHandler.NewGroup(1, 256)
	group2.AddHandlers(&roomHandler{})
	component.Register(group2)

	// add room handler

	return component
}

func createWebsocket() *cherryConnector.Component {
	connector := cherryConnector.NewWS("127.0.0.1:34590")
	component := cherryConnector.NewComponent(connector)
	component.OnConnect(func(conn facade.INetConn) {

	})

	return component
}
