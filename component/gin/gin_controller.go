package cherryGin

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/gin-gonic/gin"
)

type IController interface {
	Init(app cherryInterfaces.IApplication, engine *gin.Engine)
	Stop()
}

type BaseIController struct {
}

func (b *BaseIController) Init(_ cherryInterfaces.IApplication, _ *gin.Engine) {

}

func (b *BaseIController) Stop() {

}
