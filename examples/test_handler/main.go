package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/extend/time"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/handler"
	"github.com/cherry-game/cherry/net/message"
	"github.com/cherry-game/cherry/net/session"
	"math/rand"
	"strings"
	"time"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewApp("../config/", "local", "game-1")

	testApp.OnShutdown(
		func() {
			c := testApp.Find(cherryHandler.Name)
			if c != nil {
				cherryLogger.Debugf("--------[component = %s] is found! --------", c.Name())
			}
		},
	)

	handlerComponent := cherryHandler.NewComponent(
		cherryHandler.WithNameFunc(strings.ToLower),
	)
	//add TestHandler

	handlerGroup1 := cherryHandler.NewGroup(10, 128)
	handlerGroup1.AddHandlers(NewTestHandler())
	handlerGroup1.SetQueueHash(func(executor cherryFacade.IExecutor, queueNum int) int {
		return rand.Int() % queueNum
	})

	handlerComponent.Register(handlerGroup1)

	go func(testApp *cherry.Application) {
		//10秒后退出应用
		time.Sleep(60 * time.Second)
		testApp.Shutdown()
	}(testApp)

	testApp.Startup(
		handlerComponent,
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
	handler := m.App().Find(cherryHandler.Name).(*cherryHandler.Component)

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

		session := &cherrySession.Session{}

		msg := &cherryMessage.Message{
			Route: "game.testHandler.testLocalMethod",
			Data: []byte{
				123, 34, 110, 105, 99, 107, 110, 97, 109, 101,
				34, 58, 34, 103, 117, 101, 115, 116, 49, 54,
				50, 52, 54, 50, 48, 57, 57, 54, 55, 49, 50, 34, 125,
			},
		}

		handler.ProcessLocal(session, msg)
		//time.Sleep(time.Microsecond * 1)

		i++

		if i%100000 == 1 {
			handlerLogger.Infof("count num = %d, time = %d", i, cherryTime.Now().ToMillisecond())
		}
	}
}

func mockRequestMsg2(app cherryFacade.IApplication, handler *cherryHandler.Component) {
	for {

		session := &cherrySession.Session{}

		msg := &cherryMessage.Message{
			Route: "game.testHandler.test222",
		}

		handler.ProcessLocal(session, msg)

		//time.Sleep(time.Millisecond * 1)
	}
}

func mockEventMsg(handler *cherryHandler.Component) {
	for {
		handler.PostEvent(NewTestEvent())
		//time.Sleep(time.Millisecond * 1)
	}
}
