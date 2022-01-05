package cherryGin

import (
	cherryConst "github.com/cherry-game/cherry/const"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
)

type (
	// Component wrapper gin
	Component struct {
		cherryFacade.Component
		name       string
		httpServer *HttpServer
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
		httpServer: NewHttpServer(address, opts...),
	}
}

// Name unique components name
func (g *Component) Name() string {
	return cherryConst.HttpComponentPrefix + g.name
}

func (g *Component) Init() {
}

func (g *Component) OnAfterInit() {
	g.httpServer.SetIApplication(g.App())
	go g.httpServer.Run()
}

func (g *Component) OnBeforeStop() {
}

func (g *Component) OnStop() {
	g.httpServer.Stop()
	cherryLogger.Infof("[component = %s] has been shut down", g.Name())
}

func (g *Component) Use(middleware ...GinHandlerFunc) {
	g.httpServer.Use(middleware...)
}

func (g *Component) Register(controllers ...IController) *Component {
	g.httpServer.Register(controllers...)
	return g
}

func (g *Component) StaticFS(relativePath string, staticDir string) {
	g.httpServer.StaticFS(relativePath, staticDir)
}

func (g *Component) StaticFile(relativePath string, staticDir string) {
	g.httpServer.StaticFile(relativePath, staticDir)
}
