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
	cherryMessage "github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
	"github.com/cherry-game/cherry/net/session"
	"strings"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewApp("../profile_single/", "local", "game-1")

	testApp.OnShutdown(
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

	handlerComponent := cherryHandler.NewComponent()
	handlerComponent.SetNameFn(strings.ToLower)
	//add TestHandler

	handlerGroup1 := cherryHandler.NewGroup(10, 128)
	handlerGroup1.AddHandlers(mocks.NewTestHandler())

	handlerComponent.Register(handlerGroup1)

	dataConfigComponent := cherryDataConfig.NewComponent()

	testApp.Startup(
		handlerComponent,
		dataConfigComponent,
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
		session := &cherrySession.Session{}
		msg := &cherryMessage.Message{}

		handler.PostMessage(session, route, msg)
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

		session := &cherrySession.Session{}

		handler.PostMessage(session, route, nil)

		//time.Sleep(time.Millisecond * 1)
	}
}

func mockEventMsg(handler *cherryHandler.Component) {
	for {
		handler.PostEvent(mocks.NewTestEvent())
		//time.Sleep(time.Millisecond * 1)
	}
}
