package cherryGin

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/gin-gonic/gin"
)

type IController interface {
	PreInit(app cherryInterfaces.IApplication, engine *gin.Engine)

	Init()

	Stop()
}

type BaseIController struct {
	App    cherryInterfaces.IApplication
	Engine *gin.Engine
}

func (b *BaseIController) PreInit(app cherryInterfaces.IApplication, engine *gin.Engine) {
	b.App = app
	b.Engine = engine
}

func (b *BaseIController) Init() {

}

func (b *BaseIController) Stop() {

}
