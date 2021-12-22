package cherryGin

import (
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

func New(address string, middleware ...HandlerFunc) *Component {
	return NewHttp("http_server", address, middleware...)
}

func NewHttp(name, address string, middleware ...HandlerFunc) *Component {
	return NewHttps(name, address, "", "", middleware...)
}

func NewHttps(name, address, certFile, keyFile string, middleware ...HandlerFunc) *Component {
	component := NewWithOptions(name, address, Options{
		CertFile: certFile,
		KeyFile:  keyFile,
	})

	//add default middleware
	component.Use(middleware...)

	return component
}

func NewWithOptions(name string, address string, options Options) *Component {
	if address == "" {
		cherryLogger.Warnf("[%s] no set listener address.", name)
		return nil
	}

	httpServer := NewHttpServer(address)
	httpServer.SetOptions(options)

	return &Component{
		name:       name,
		httpServer: httpServer,
	}
}

func (g *Component) Use(middleware ...HandlerFunc) {
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

// Name unique components name
func (g *Component) Name() string {
	return g.name
}

func (g *Component) Init() {
}

func (g *Component) OnAfterInit() {
	g.httpServer.SetIApplication(g.App())
	g.httpServer.Run(true)
}

func (g *Component) OnBeforeStop() {
}

func (g *Component) OnStop() {
	g.httpServer.Stop()
	cherryLogger.Infof("[component = %s] has been shut down", g.name)
}
