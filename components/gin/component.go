package cherryGin

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/gin-gonic/gin"
)

const (
	NamePrefix = "gin_component_"
)

type (
	// Component wrapper gin
	Component struct {
		cfacade.Component
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
	return NamePrefix + g.name
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
	clog.Infof("[component = %s] has been shut down", g.Name())
}

func (g *Component) Register(controllers ...IController) *Component {
	g.HttpServer.Register(controllers...)
	return g
}

func (g *Component) Engine() *gin.Engine {
	return g.HttpServer.Engine
}
