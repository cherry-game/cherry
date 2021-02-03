package cherryGin

import (
	"github.com/cherry-game/cherry/interfaces"
	"github.com/gin-gonic/gin"
)

type GinHandlerFunc func(ctx *GinContext)

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

func (b *BaseIController) Any(relativePath string, handlers ...GinHandlerFunc) {
	for _, handler := range handlers {
		b.Engine.Any(relativePath, b.BindHandle(handler))
	}
}

func (b *BaseIController) GET(relativePath string, handlers ...GinHandlerFunc) {
	for _, handler := range handlers {
		b.Engine.GET(relativePath, b.BindHandle(handler))
	}
}

func (b *BaseIController) POST(relativePath string, handlers ...GinHandlerFunc) {
	for _, handler := range handlers {
		b.Engine.POST(relativePath, b.BindHandle(handler))
	}
}

func (b *BaseIController) BindHandle(handler func(ctx *GinContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := new(GinContext)
		context.Context = c
		handler(context)
	}
}
