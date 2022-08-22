package main

import (
	"github.com/cherry-game/cherry"
	ctime "github.com/cherry-game/cherry/extend/time"
	clog "github.com/cherry-game/cherry/logger"
	cconnector "github.com/cherry-game/cherry/net/connector"
	chandler "github.com/cherry-game/cherry/net/handler"
	cmsg "github.com/cherry-game/cherry/net/message"
	cserializer "github.com/cherry-game/cherry/net/serializer"
	csession "github.com/cherry-game/cherry/net/session"
	"testing"
	"time"
)

func TestGame1(t *testing.T) {
	//c1.Test()
}

func TestPostMessage(t *testing.T) {
	app := cherry.NewApp("../config/", "local", "game-1")
	app.SetSerializer(cserializer.NewJSON())

	handlerComponent := createHandler()

	go func() {
		time.Sleep(5 * time.Second)

		session := &csession.Session{}

		msg := &cmsg.Message{
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

			handlerComponent.ProcessLocal(session, msg)
			//time.Sleep(time.Microsecond * 1)

			i++

			if i%100000 == 1 {
				clog.Infof("count num = %d, time = %d", i, ctime.Now().ToMillisecond())
			}
		}

	}()

	app.Startup(
		handlerComponent,
		cconnector.NewWS("127.0.0.1:34590"),
	)
}

func createHandler() *chandler.Component {
	component := chandler.NewComponent()

	group1 := chandler.NewGroup(1, 256)
	group1.AddHandlers(&userHandler{})
	component.Register(group1)

	group2 := chandler.NewGroup(1, 256)
	group2.AddHandlers(&roomHandler{})
	component.Register(group2)

	// add room handler

	return component
}
