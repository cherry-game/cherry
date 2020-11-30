package main

import (
	"github.com/cherry-game/cherry"
	cherryQueue "github.com/cherry-game/cherry/components/queue"
	"github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/examples/test1/mocks"
	"github.com/cherry-game/cherry/handler"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net"
	"github.com/cherry-game/cherry/session"
	"strings"
	"time"
)

func main() {

	app(cherry.DefaultAppParameters())
}

func app(configPath, profileName, nodeId string) {
	testApp := cherry.NewApp(configPath, profileName, nodeId)

	defer testApp.Shutdown(func() {
		c := testApp.Find(cherryConst.HandlerComponent)
		if c != nil {
			cherryLogger.Debugf("--------[component = %s] is find! --------", c.Name())
		}
	})

	handlers := cherryHandler.NewComponent()
	handlers.SetNameFunc(strings.ToLower)
	handlers.BeforeFilter(func(msg cherryHandler.UnhandledMessage) bool {
		cherryLogger.Infof("test before filter....")
		return true
	})
	handlers.AfterFilter(func(msg cherryHandler.UnhandledMessage) bool {
		cherryLogger.Infof("test after filter....")
		return true
	})
	//add TestHandler
	handlers.Register(mocks.NewTestHandler())

	testApp.Startup(
		handlers,
		cherryQueue.NewQueue(),
	)

	go mockRequestMsg1(handlers)
	go mockRequestMsg2(handlers)
	go mockEventMsg(handlers)
}

func mockRequestMsg1(handler *cherryHandler.HandlerComponent) {
	for {
		route := cherryNet.NewByName("game.testHandler.test11111")
		handler.InHandle(route, &cherrySession.Session{}, nil)
		time.Sleep(time.Millisecond * 5)
	}
}

func mockRequestMsg2(handler *cherryHandler.HandlerComponent) {
	for {
		route := cherryNet.NewByName("game.testHandler.test222")
		handler.InHandle(route, &cherrySession.Session{}, nil)
		time.Sleep(time.Millisecond * 5)
	}
}

func mockEventMsg(handler *cherryHandler.HandlerComponent) {
	for {
		handler.PostEvent(mocks.NewTestEvent())
		time.Sleep(time.Millisecond * 5)
	}
}
