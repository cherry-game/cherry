package cherryGin

import (
	cherryInterfaces "github.com/cherry-game/cherry/interfaces"
	"github.com/gin-gonic/gin"
)

type IController interface {
	Init(app cherryInterfaces.AppContext, engine *gin.Engine)
}
