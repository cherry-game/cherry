package main

import (
	cherryConst "github.com/cherry-game/cherry/const"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryHandler "github.com/cherry-game/cherry/net/handler"
	"time"
)

type MockClientComponent struct {
	cherryFacade.Component
	handlerComponent *cherryHandler.Component
}

func (m *MockClientComponent) Name() string {
	return "mock_client_component"
}

func (m *MockClientComponent) Init() {

}

func (m *MockClientComponent) OnAfterInit() {
	m.handlerComponent, _ = m.App().Find(cherryConst.HandlerComponent).(*cherryHandler.Component)
	if m.handlerComponent == nil {
		panic("handler component not found.")
	}

	go func() {

		for {
			cherryLogger.Debugf("send msg")

			time.Sleep(1 * time.Second)
		}
	}()
}
