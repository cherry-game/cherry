package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test1_handler/mocks"
	"github.com/cherry-game/cherry/data_config"
	"github.com/cherry-game/cherry/handler"
	"github.com/cherry-game/cherry/logger"
)

func main() {
	app()
}

func app() {
	testApp := cherry.NewDefaultApp()

	defer testApp.Shutdown()

	handlers := cherryHandler.NewComponent()

	dataConfig := cherryDataConfig.NewComponent()
	dataConfig.Register(new(mocks.DropConfig))

	testApp.Startup(
		handlers,
		dataConfig,
	)
	cherryLogger.Infow("test", "key", "itemId", "value", 2)
}
