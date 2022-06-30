package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/component/data_config"
	cmini "github.com/cherry-game/cherry/component/mini"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	chandler "github.com/cherry-game/cherry/net/handler"
	"time"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewApp("../config/", "local", "game-1")

	handlers := chandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(&DropList, &DropOne)

	go func(testApp *cherry.Application) {
		//120秒后退出应用
		time.Sleep(120 * time.Second)
		testApp.Shutdown()
	}(testApp)

	mockComponent := cmini.New(
		"mock",
		cmini.WithAfterInit(func(app cfacade.IApplication) {
			go getDropConfig(app)
		}),
	)

	mock1Component1 := cmini.New(
		"mock1",
		cmini.WithInitFunc(func(_ cfacade.IApplication) {
			clog.Info("call init func")
		}),
		cmini.WithAfterInit(func(_ cfacade.IApplication) {
			clog.Info("call after init func")
		}),
		cmini.WithBeforeStop(func(_ cfacade.IApplication) {
			clog.Info("call before stop func")
		}),
		cmini.WithStop(func(_ cfacade.IApplication) {
			clog.Info("call stop func")
		}),
	)

	testApp.Startup(
		handlers,
		dataConfig,
		mockComponent,
		mock1Component1,
	)
}

func getDropConfig(_ cfacade.IApplication) {

	time.Sleep(5 * time.Second)

	for {
		clog.Infof("DropOneConfig %p, %v", &DropOne, DropOne)

		x1 := DropList.Get(1011)
		clog.Infof("DropConfig %p, %v", x1, x1)

		itemTypeList := DropList.GetItemTypeList(3)
		clog.Infof("DropConfig %p, %v", itemTypeList, itemTypeList)

		time.Sleep(500 * time.Millisecond)
	}
}
