package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/data_config"
	cherryFacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/handler"
	"time"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewApp("../profile_split/", "local", "game-1")

	handlers := cherryHandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(&DropList, &DropOne)

	cherryLogger.Infow("test", "key", "itemId", "value", 2)

	go getDropConfig(testApp)

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
	go getDropConfig(m.App())
}

func getDropConfig(_ cherryFacade.IApplication) {
	for {
		x1 := DropList.Get(1011)
		//cherryLogger.Info(x1)
		cherryLogger.Warnf("%p, %v", x1, x1)

		cherryLogger.Warnf("%p, %v", &DropOne, DropOne)

		itemTypeList := DropList.GetItemTypeList(3)
		cherryLogger.Warnf("%p, %v", itemTypeList, itemTypeList)

		time.Sleep(10 * time.Millisecond)
	}
}
