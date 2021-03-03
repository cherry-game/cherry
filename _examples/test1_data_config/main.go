package main

import (
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/_examples/test1_handler/mocks"
	cherryConst "github.com/cherry-game/cherry/const"
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

	dataConfig.Register("dropConfig", "dropOneConfig")

	testApp.Startup(
		handlers,
		dataConfig,
	)
	cherryLogger.Infow("test", "key", "itemId", "value", 2)

	go getDropConfig(testApp)
}

func getDropConfig(app *cherry.Application) {
	component := app.Find(cherryConst.DataConfigComponent).(*cherryDataConfig.DataConfigComponent)

	var list []mocks.DropConfig
	component.Get("dropConfig", &list)
	cherryLogger.Warnf("%p", list)

	var list1 []mocks.DropConfig
	component.Get("dropConfig", &list1)
	cherryLogger.Warnf("%p", list1)

	var one mocks.DropOneConfig
	component.Get("dropOneConfig", &one)
	cherryLogger.Warnf("%p", &one)
}
