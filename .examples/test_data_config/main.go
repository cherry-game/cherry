package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/.examples/test1/mocks"
	"github.com/cherry-game/cherry/data_config"
	"github.com/cherry-game/cherry/handler"
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
}
