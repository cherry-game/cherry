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
		cherryComponents.NewQueue(),
	)

	go mockRequestMsg1(testApp)
	go mockRequestMsg2(testApp)
	go mockEventMsg(testApp)
}

func mockRequestMsg1(app *cherry.Application) {
	for {
		route := cherryNet.NewByName("web.testHandler.test11111")
		app.Handlers().InHandle(route, &cherrySession.Session{}, nil)
		time.Sleep(time.Millisecond * 5)
	}
}

func mockRequestMsg2(app *cherry.Application) {
	for {
		route := cherryNet.NewByName("web.testHandler.test222")
		app.Handlers().InHandle(route, &cherrySession.Session{}, nil)
		time.Sleep(time.Millisecond * 5)
	}
}

func mockEventMsg(app *cherry.Application) {
	for {
		app.PostEvent(mocks.NewTestEvent())
		time.Sleep(time.Millisecond * 5)
	}
}
