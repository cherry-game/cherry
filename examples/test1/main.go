package main

import (
	"flag"
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/components"
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
	var configPath, profileName, nodeId string

	flag.StringVar(&configPath, "path", "./config", "-path=~/git/cherry/examples/config")
	flag.StringVar(&profileName, "profile", "local", "-profile=local")
	flag.StringVar(&nodeId, "node", "web-1", "-node=web-1")
	flag.Parse()

	app(configPath, profileName, nodeId)
}

func app(configPath, profileName, nodeId string) {
	testApp := cherry.CreateApp(configPath, profileName, nodeId)

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
		cherryComponents.NewQueue(),
	)

	go mockRequestMsg1(handlers)
	go mockRequestMsg2(handlers)
	go mockEventMsg(handlers)
}

func mockRequestMsg1(handler *cherryHandler.HandlerComponent) {
	for {
		route := cherryNet.NewByName("web.testHandler.test11111")
		handler.InHandle(route, &cherrySession.Session{}, nil)
		time.Sleep(time.Millisecond * 5)
	}
}

func mockRequestMsg2(handler *cherryHandler.HandlerComponent) {
	for {
		route := cherryNet.NewByName("web.testHandler.test222")
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
