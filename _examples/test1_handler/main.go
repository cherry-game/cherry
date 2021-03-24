package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test1_handler/mocks"
	"github.com/cherry-game/cherry/component/queue"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/data_config"
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
				cherryLogger.Debugf("--------[component = %s] is find! --------", c.Name())
			}
		},
		func() {
			cherryLogger.DefaultLogger().Sync()

			handlerLogger := cherryLogger.NewLogger("test_handler")
			handlerLogger.Sync()
		},
	)

	handlers := cherryHandler.NewComponent()

	handlers.SetNameFn(strings.ToLower)

	handlers.BeforeFilter(func(msg *cherryHandler.UnhandledMessage) bool {
		cherryLogger.Debug("test before filter.... ")
		return false
	})

	handlers.AfterFilter(func(msg *cherryHandler.UnhandledMessage) bool {
		cherryLogger.Debug("test after filter....")
		return true
	})

	//add TestHandler
	handlers.Registers(mocks.NewTestHandler())
	dataConfig := cherryDataConfig.NewComponent()

	testApp.Startup(
		handlers,
		dataConfig,
		cherryQueue.NewQueue(),
	)

	go mockRequestMsg1(handlers)
	go mockRequestMsg2(handlers)
	//go mockEventMsg(handlers)
}

func mockRequestMsg1(handler *cherryHandler.HandlerComponent) {
	handlerLogger := cherryLogger.NewLogger("test_handler")

	for {
		route := cherryRoute.NewByName("game.testHandler.test11111")

		handlerLogger.Infow("", "route", route.String())

		msg := &cherryHandler.UnhandledMessage{
			Session: &cherrySession.Session{},
			Route:   route,
			Msg:     nil,
		}

		handler.DoHandle(msg)
		//time.Sleep(time.Microsecond * 1)
	}
}

func mockRequestMsg2(handler *cherryHandler.HandlerComponent) {
	handlerLogger := cherryLogger.NewLogger("test_handler")

	for {
		route := cherryRoute.NewByName("game.testHandler.test222")

		handlerLogger.Infow("", "route", route.String())

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
