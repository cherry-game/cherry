package cherryGin

import (
	"context"
	"github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type (
	ComponentOptions struct {
		ReadTimeout       time.Duration // http server parameter
		ReadHeaderTimeout time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		MaxHeaderBytes    int
		Address           string
		CertFile          string
		KeyFile           string
	}

	// Component wrapper gin
	Component struct {
		cherryFacade.Component
		name        string
		engine      *gin.Engine
		server      *http.Server
		options     ComponentOptions
		controllers []IController
	}
)

func New(address string, middleware ...gin.HandlerFunc) *Component {
	return NewHttp("http_server", address, middleware...)
}

func NewHttp(name, address string, middleware ...gin.HandlerFunc) *Component {
	return NewHttps(name, address, "", "", middleware...)
}

func NewHttps(name, address, certFile, keyFile string, middleware ...gin.HandlerFunc) *Component {
	component := NewWithOptions(name, ComponentOptions{
		Address:  address,
		CertFile: certFile,
		KeyFile:  keyFile,
	})

	//add default middleware
	component.engine.Use(middleware...)

	return component
}

func NewWithOptions(name string, options ComponentOptions) *Component {
	return &Component{
		name:    name,
		engine:  gin.New(),
		server:  &http.Server{},
		options: options,
	}
}

func (g *Component) Register(controllers ...IController) *Component {
	for _, controller := range controllers {
		g.controllers = append(g.controllers, controller)
	}
	return g
}

func (g *Component) GetEngine() *gin.Engine {
	return g.engine
}

// Name unique components name
func (g *Component) Name() string {
	return g.name
}

func (g *Component) Init() {
	if g.options.Address == "" {
		cherryLogger.Warnf("[%s] no set address value.", g.name)
		return
	}

	g.server.Handler = g.engine
	g.server.Addr = g.options.Address

	if g.options.ReadTimeout > 0 {
		g.server.ReadTimeout = g.options.ReadTimeout
	}

	if g.options.ReadHeaderTimeout > 0 {
		g.server.ReadHeaderTimeout = g.options.ReadHeaderTimeout
	}

	if g.options.WriteTimeout > 0 {
		g.server.WriteTimeout = g.options.WriteTimeout
	}

	if g.options.IdleTimeout > 0 {
		g.server.IdleTimeout = g.options.IdleTimeout
	}

	if g.options.MaxHeaderBytes > 0 {
		g.server.MaxHeaderBytes = g.options.MaxHeaderBytes
	}

	for _, controller := range g.controllers {
		controller.PreInit(g.App(), g.engine)
		controller.Init()
	}

	go func() {
		var err error

		if g.options.CertFile != "" && g.options.KeyFile != "" {
			cherryLogger.Infof("[%s] -> https init. address = %s, certFile = %s, keyFile = %s",
				g.name, g.options.Address, g.options.CertFile, g.options.KeyFile)
			err = g.server.ListenAndServeTLS(g.options.CertFile, g.options.KeyFile)
		} else {
			cherryLogger.Infof("[%s] -> http init. address = %s", g.name, g.options.Address)
			err = g.server.ListenAndServe()
		}

		if err != nil {
			cherryLogger.Infof("[%s] run error = %s", g.name, err)
		}
	}()
}

func (g *Component) OnAfterInit() {

}

func (g *Component) OnBeforeStop() {
	for _, controller := range g.controllers {
		controller.Stop()
	}
}

func (g *Component) OnStop() {
	err := g.server.Shutdown(context.Background())
	cherryLogger.Infof("[%s] shutdown gin component on %s", g.name, g.options.Address)

	if err != nil {
		cherryLogger.Info(err.Error())
	}
}
