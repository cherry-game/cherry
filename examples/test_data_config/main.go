package main

import (
	"github.com/cherry-game/cherry"
	cherryDataConfig "github.com/cherry-game/cherry/data-config"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	chandler "github.com/cherry-game/cherry/net/handler"
	"time"
)

func main() {
	testApp := cherry.NewApp("../config/", "local", "game-1")

	handlers := chandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(&DropList, &DropOne)

	go func(testApp *cherry.Application) {
		//120秒后退出应用
		getDropConfig(testApp)
		testApp.Shutdown()
	}(testApp)

	testApp.Startup(
		handlers,
		dataConfig,
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
