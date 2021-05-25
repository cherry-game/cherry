package cherryGin

import (
	"github.com/cherry-game/cherry/facade"
	"github.com/gin-gonic/gin"
)

type GinHandlerFunc func(ctx *GinContext)

type IController interface {
	PreInit(app cherryFacade.IApplication, engine *gin.Engine)

	Init()

	Stop()
}

type BaseController struct {
	App    cherryFacade.IApplication
	Engine *gin.Engine
}

func (b *BaseController) PreInit(app cherryFacade.IApplication, engine *gin.Engine) {
	b.App = app
	b.Engine = engine
}

func (b *BaseController) Init() {

}

func (b *BaseController) Stop() {

}

func (b *BaseController) Any(relativePath string, handlers ...GinHandlerFunc) {
	for _, handler := range handlers {
		b.Engine.Any(relativePath, b.BindHandle(handler))
	}
}

func (b *BaseController) GET(relativePath string, handlers ...GinHandlerFunc) {
	for _, handler := range handlers {
		b.Engine.GET(relativePath, b.BindHandle(handler))
	}
}

func (b *BaseController) POST(relativePath string, handlers ...GinHandlerFunc) {
	for _, handler := range handlers {
		b.Engine.POST(relativePath, b.BindHandle(handler))
	}
}

func (b *BaseController) BindHandle(handler func(ctx *GinContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		context := new(GinContext)
		context.Context = c
		handler(context)
	}
}
