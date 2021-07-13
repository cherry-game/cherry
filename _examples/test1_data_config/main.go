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
	testApp := cherry.NewApp("../profile_single/", "local", "game-1")

	handlers := cherryHandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(&DropList, &DropOne)

	go func(testApp *cherry.Application) {
		//120秒后退出应用
		time.Sleep(120 * time.Second)
		testApp.Shutdown()
	}(testApp)

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

	time.Sleep(5 * time.Second)

	for {
		cherryLogger.Infof("DropOneConfig %p, %v", &DropOne, DropOne)

		x1 := DropList.Get(1011)
		cherryLogger.Infof("DropConfig %p, %v", x1, x1)

		itemTypeList := DropList.GetItemTypeList(3)
		cherryLogger.Infof("DropConfig %p, %v", itemTypeList, itemTypeList)

		time.Sleep(500 * time.Millisecond)
	}
}
