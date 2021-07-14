package cherryGin

import (
	"context"
	"github.com/cherry-game/cherry/extend/file"
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
	Options struct {
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
		*gin.Engine
		name        string
		server      *http.Server
		options     Options
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
	component := NewWithOptions(name, Options{
		Address:  address,
		CertFile: certFile,
		KeyFile:  keyFile,
	})

	//add default middleware
	component.Use(middleware...)

	return component
}

func NewWithOptions(name string, options Options) *Component {
	return &Component{
		name:    name,
		Engine:  gin.New(),
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

func (g *Component) StaticFS(relativePath string, staticDir string) {
	dir, ok := cherryFile.JudgePath(staticDir)
	if !ok {
		cherryLogger.Errorf("static dir path not found. staticDir = %s", staticDir)
		return
	}
	g.Engine.StaticFS(relativePath, http.Dir(dir))
}

func (g *Component) StaticFile(relativePath string, staticDir string) {
	dir, ok := cherryFile.JudgePath(staticDir)
	if !ok {
		cherryLogger.Errorf("static dir path not found. staticDir = %s", staticDir)
		return
	}
	g.Engine.StaticFile(relativePath, dir)
}

// Name unique components name
func (g *Component) Name() string {
	return g.name
}

func (g *Component) Init() {
}

func (g *Component) OnAfterInit() {
	if g.options.Address == "" {
		cherryLogger.Warnf("[%s] no set address value.", g.name)
		return
	}

	g.server.Handler = g
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
		controller.PreInit(g.App(), g.Engine)
		controller.Init()
	}

	go func() {
		var err error

		if g.options.CertFile != "" && g.options.KeyFile != "" {
			cherryLogger.Infof("[component = %s] https init. https://%s, certFile = %s, keyFile = %s",
				g.name, g.options.Address, g.options.CertFile, g.options.KeyFile)
			err = g.server.ListenAndServeTLS(g.options.CertFile, g.options.KeyFile)
		} else {
			cherryLogger.Infof("[component = %s] http init. http://%s", g.name, g.options.Address)
			err = g.server.ListenAndServe()
		}

		if err != http.ErrServerClosed {
			cherryLogger.Infof("[component = %s] run error = %s", g.name, err)
		}
	}()
}

func (g *Component) OnBeforeStop() {
	for _, controller := range g.controllers {
		controller.Stop()
	}
}

func (g *Component) OnStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := g.server.Shutdown(ctx); err != nil {
		cherryLogger.Info(err.Error())
	}

	cherryLogger.Infof("[component = %s] shutdown gin component on %s", g.name, g.options.Address)
}
