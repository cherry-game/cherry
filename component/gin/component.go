package cherryGin

import (
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/gin-gonic/gin"
)

type (
	// Component wrapper gin
	Component struct {
		cherryFacade.Component
		*HttpServer
		name string
	}
)

func NewHttp(name, address string) *Component {
	return New(name, address)
}

func NewHttps(name, address, certFile, keyFile string) *Component {
	return New(
		name,
		address,
		WithCert(certFile, keyFile),
	)
}

func New(name string, address string, opts ...OptionFunc) *Component {
	return &Component{
		name:       name,
		HttpServer: NewHttpServer(address, opts...),
	}
}

// Name unique components name
func (g *Component) Name() string {
	return cherryConst.HttpComponentPrefix + g.name
}

func (g *Component) Init() {
}

func (g *Component) OnAfterInit() {
	g.SetIApplication(g.App())
	go g.Run()
}

func (g *Component) OnBeforeStop() {
}

func (g *Component) OnStop() {
	g.Stop()
	cherryLogger.Infof("[component = %s] has been shut down", g.Name())
}

func (g *Component) Register(controllers ...IController) *Component {
	g.HttpServer.Register(controllers...)
	return g
}

func (g *Component) Engine() *gin.Engine {
	return g.HttpServer.Engine
}
