package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test1_handler/mocks"
	"github.com/cherry-game/cherry/component/data_config"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/time"
	cherryFacade "github.com/cherry-game/cherry/facade"
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
	testApp := cherry.NewApp("../profile_single/", "local", "game-1")

	defer testApp.OnShutdown(
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
		&mockComponent{},
	)
}

type mockComponent struct {
	cherryFacade.Component
}

func (m *mockComponent) Name() string {
	return "mock_component"
}

func (m *mockComponent) OnAfterInit() {
	handler := m.App().Find(cherryConst.HandlerComponent).(*cherryHandler.Component)

	go mockRequestMsg1(handler)
	//go mockRequestMsg2(handler)
	//go mockEventMsg(handler)
}

func mockRequestMsg1(handler *cherryHandler.Component) {
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

		if i%100000 == 1 {
			handlerLogger.Infof("count num = %d, time = %d", i, cherryTime.Now().ToMillisecond())
		}
	}
}

func mockRequestMsg2(handler *cherryHandler.Component) {
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

func mockEventMsg(handler *cherryHandler.Component) {
	for {
		handler.PostEvent(mocks.NewTestEvent())
		//time.Sleep(time.Millisecond * 1)
	}
}
