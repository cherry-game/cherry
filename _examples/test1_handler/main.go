package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test1_handler/mocks"
	"github.com/cherry-game/cherry/component/data_config"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/extend/time"
	cherryFacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	cherryAgent "github.com/cherry-game/cherry/net/agent"
	"github.com/cherry-game/cherry/net/handler"
	cherryMessage "github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/route"
	cherrySerializer "github.com/cherry-game/cherry/net/serializer"
	"github.com/cherry-game/cherry/net/session"
	"math/rand"
	"strings"
	"time"
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
	)

	handlerComponent := cherryHandler.NewComponent()
	handlerComponent.SetNameFn(strings.ToLower)
	//add TestHandler

	handlerGroup1 := cherryHandler.NewGroup(30, 128)
	handlerGroup1.AddHandlers(mocks.NewTestHandler())
	handlerGroup1.SetQueueHash(func(executor cherryHandler.IExecutor, queueNum int) int {
		return rand.Int() % queueNum
	})

	handlerComponent.Register(handlerGroup1)

	dataConfigComponent := cherryDataConfig.NewComponent()

	go func(testApp *cherry.Application) {
		//10秒后退出应用
		time.Sleep(10 * time.Second)
		testApp.Shutdown()
	}(testApp)

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

	go mockRequestMsg1(m.App(), handler)
	//go mockRequestMsg2(handler)
	//go mockEventMsg(handler)
}

func mockRequestMsg1(app cherryFacade.IApplication, handler *cherryHandler.Component) {
	handlerLogger := cherryLogger.NewLogger("test_handler")
	i := 0

	handlerLogger.Info(cherryTime.Now().ToMillisecond())

	for {

		if app.Running() == false {
			return
		}

		route := cherryRoute.NewByName("game.testHandler.testLocalMethod")

		agent := &cherryAgent.Agent{
			Options: cherryAgent.Options{
				Serializer: cherrySerializer.NewJSON(),
			},
			Session: &cherrySession.Session{},
		}

		msg := &cherryMessage.Message{}

		handler.PostMessage(agent, route, msg)
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

		agent := &cherryAgent.Agent{
			Session: &cherrySession.Session{},
		}

		handler.PostMessage(agent, route, nil)

		//time.Sleep(time.Millisecond * 1)
	}
}

func mockEventMsg(handler *cherryHandler.Component) {
	for {
		handler.PostEvent(mocks.NewTestEvent())
		//time.Sleep(time.Millisecond * 1)
	}
}
