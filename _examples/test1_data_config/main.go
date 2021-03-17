package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/data_config"
	"github.com/cherry-game/cherry/handler"
	"github.com/cherry-game/cherry/logger"
	"time"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewDefaultApp()

	defer testApp.Shutdown()

	handlers := cherryHandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()

	dataConfig.Register(&DropConfigList, &DropOneConfig)

	testApp.Startup(
		handlers,
		dataConfig,
	)
	cherryLogger.Infow("test", "key", "itemId", "value", 2)

	go getDropConfig(testApp)
}

func getDropConfig(_ *cherry.Application) {

	for {

		x1 := DropConfigList.Get(1011)
		//cherryLogger.Info(x1)
		cherryLogger.Warnf("%p, %v", x1, x1)

		cherryLogger.Warnf("%p, %v", &DropOneConfig, DropOneConfig)
		time.Sleep(1 * time.Second)
	}
}
