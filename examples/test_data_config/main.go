package main

import (
	"github.com/cherry-game/cherry"
	cherryDataConfig "github.com/cherry-game/cherry/components/data-config"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"time"
)

func main() {
	testApp := cherry.NewApp(
		"./examples/config/profile-local.json",
		"game-1",
		false,
		cherry.Standalone,
	)

	dataConfig := cherryDataConfig.New()
	dataConfig.Register(&DropList, &DropOne)
	testApp.Register(dataConfig)

	go func(testApp *cherry.Application) {
		//120秒后退出应用
		getDropConfig(testApp)
		testApp.Shutdown()
	}(testApp)

	testApp.Startup()
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
