package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/.examples/test1/mocks"
	"github.com/cherry-game/cherry/component/queue"
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/data_config"
	"github.com/cherry-game/cherry/handler"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
	"strings"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewDefaultApp()

	defer testApp.Shutdown(func() {
		c := testApp.Find(cherryConst.HandlerComponent)
		if c != nil {
			cherryLogger.Debugf("--------[component = %s] is find! --------", c.Name())
		}
	})

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
	dataConfig.Register(new(mocks.DropConfig))

	testApp.Startup(
		handlers,
		dataConfig,
		cherryQueue.NewQueue(),
	)

	go mockRequestMsg1(handlers)
	go mockRequestMsg2(handlers)
	go mockEventMsg(handlers)
}

func mockRequestMsg1(handler *cherryHandler.HandlerComponent) {
	for {
		route := cherryRoute.NewByName("game.testHandler.test11111")

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
