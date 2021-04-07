package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test1_handler/mocks"
	"github.com/cherry-game/cherry/component/data_config"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
	"strings"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewDefaultApp()

	defer testApp.Shutdown(
		func() {
			c := testApp.Find(cherryConst.HandlerComponent)
			if c != nil {
				cherryLogger.Debugf("--------[component = %s] is found! --------", c.Name())
			}
		},
		func() {
			cherryLogger.DefaultLogger().Sync()
			handlerLogger := cherryLogger.NewLogger("test_handler")
			handlerLogger.Sync()

			timeLogger := cherryLogger.NewLogger("test_handler")
			timeLogger.Info(cherryTime.Now().ToMillisecond())
		},
	)

	handlers := cherryHandler.NewComponent()
	handlers.SetNameFn(strings.ToLower)

	//add TestHandler
	handlers.Registers(mocks.NewTestHandler())

	dataConfig := cherryDataConfig.NewComponent()

	testApp.Startup(
		handlers,
		dataConfig,
	)

	go mockRequestMsg1(handlers)
	//go mockRequestMsg2(handlers)
	//go mockEventMsg(handlers)
}

func mockRequestMsg1(handler *cherryHandler.HandlerComponent) {
	handlerLogger := cherryLogger.NewLogger("test_handler")
	i := 0

	handlerLogger.Info(cherryTime.Now().ToMillisecond())

	for {
		route := cherryRoute.NewByName("game.testHandler.testLocalMethod")

		msg := &cherryHandler.UnhandledMessage{
			Session: &cherrySession.Session{},
			Route:   route,
			Msg:     nil,
		}

		handler.DoHandle(msg)
		//time.Sleep(time.Microsecond * 1)

		i++

		if i%1000000 == 0 {
			handlerLogger.Infof("count num = %d, time = %d", i, cherryTime.Now().ToMillisecond())
			break
		}
	}
}

func mockRequestMsg2(handler *cherryHandler.HandlerComponent) {
	for {
		route := cherryRoute.NewByName("game.testHandler.test222")

		msg := &cherryHandler.UnhandledMessage{
			Session: &cherrySession.Session{},
			Route:   route,
			Msg:     nil,
		}

		handler.DoHandle(msg)
		//time.Sleep(time.Millisecond * 1)
	}
}

func mockEventMsg(handler *cherryHandler.HandlerComponent) {
	for {
		handler.PostEvent(mocks.NewTestEvent())
		//time.Sleep(time.Millisecond * 1)
	}
}
