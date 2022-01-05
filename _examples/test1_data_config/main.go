package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/data_config"
	cherryMini "github.com/cherry-game/cherry/component/mini"
	cf "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/handler"
	"time"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewApp("../config/", "local", "game-1")

	handlers := cherryHandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(&DropList, &DropOne)

	go func(testApp *cherry.Application) {
		//120秒后退出应用
		time.Sleep(120 * time.Second)
		testApp.Shutdown()
	}(testApp)

	mockComponent := cherryMini.New("mock")
	mockComponent.SetAfterInit(func(app cf.IApplication) {
		go getDropConfig(app)
	})

	testApp.Startup(
		handlers,
		dataConfig,
		mockComponent,
	)
}

func getDropConfig(_ cf.IApplication) {

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
