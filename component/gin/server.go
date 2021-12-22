package cherryGin

import (
	"context"
	cherryFile "github.com/cherry-game/cherry/extend/file"
	cherryFacade "github.com/cherry-game/cherry/facade"
	cherryLogger "github.com/cherry-game/cherry/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type (
	// Options http server parameter
	Options struct {
		ReadTimeout       time.Duration
		ReadHeaderTimeout time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		MaxHeaderBytes    int
		CertFile          string
		KeyFile           string
	}

	HttpServer struct {
		appContext cherryFacade.IApplication
		*gin.Engine
		server      *http.Server
		options     Options
		controllers []IController
	}
)

func init() {
	SetMode(gin.ReleaseMode)
}

func SetMode(value string) {
	gin.SetMode(value)
}

func NewHttpServer(address string, middleware ...HandlerFunc) *HttpServer {
	return NewHttpsServer(address, "", "", middleware...)
}

func NewHttpsServer(address, certFile, keyFile string, middleware ...HandlerFunc) *HttpServer {
	httpServer := &HttpServer{}
	httpServer.Engine = gin.New()
	httpServer.Use(middleware...)
	httpServer.server = &http.Server{
		Addr:    address,
		Handler: httpServer.Engine,
	}
	httpServer.options = Options{
		CertFile: certFile,
		KeyFile:  keyFile,
	}

	return httpServer
}

func (p *HttpServer) SetOptions(options Options) {
	p.options = options
}

func (p *HttpServer) Use(middleware ...HandlerFunc) {
	p.Engine.Use(BindHandlers(middleware)...)
}

func (p *HttpServer) SetIApplication(appContext cherryFacade.IApplication) {
	p.appContext = appContext
}

func (p *HttpServer) Register(controllers ...IController) *HttpServer {
	for _, controller := range controllers {
		p.controllers = append(p.controllers, controller)
	}
	return p
}

func (p *HttpServer) StaticFS(relativePath string, staticDir string) {
	dir, ok := cherryFile.JudgePath(staticDir)
	if !ok {
		cherryLogger.Errorf("static dir path not found. staticDir = %s", staticDir)
		return
	}
	p.Engine.StaticFS(relativePath, http.Dir(dir))
}

func (p *HttpServer) StaticFile(relativePath string, staticDir string) {
	dir, ok := cherryFile.JudgePath(staticDir)
	if !ok {
		cherryLogger.Errorf("static dir path not found. staticDir = %s", staticDir)
		return
	}
	p.Engine.StaticFile(relativePath, dir)
}

func (p *HttpServer) Run(async ...bool) {
	if p.server.Addr == "" {
		cherryLogger.Warn("no set listener address.")
		return
	}

	if p.options.ReadTimeout > 0 {
		p.server.ReadTimeout = p.options.ReadTimeout
	}

	if p.options.ReadHeaderTimeout > 0 {
		p.server.ReadHeaderTimeout = p.options.ReadHeaderTimeout
	}

	if p.options.WriteTimeout > 0 {
		p.server.WriteTimeout = p.options.WriteTimeout
	}

	if p.options.IdleTimeout > 0 {
		p.server.IdleTimeout = p.options.IdleTimeout
	}

	if p.options.MaxHeaderBytes > 0 {
		p.server.MaxHeaderBytes = p.options.MaxHeaderBytes
	}

	for _, controller := range p.controllers {
		controller.PreInit(p.appContext, p.Engine)
		controller.Init()
	}

	asyncFlag := false
	if len(async) > 0 {
		asyncFlag = async[0]
	}

	if asyncFlag {
		go p.listener()
	} else {
		p.listener()
	}
}

func (p *HttpServer) listener() {
	var err error
	if p.options.CertFile != "" && p.options.KeyFile != "" {
		cherryLogger.Infof("https run. https://%s, certFile = %s, keyFile = %s",
			p.server.Addr, p.options.CertFile, p.options.KeyFile)
		err = p.server.ListenAndServeTLS(p.options.CertFile, p.options.KeyFile)
	} else {
		cherryLogger.Infof("http run. http://%s", p.server.Addr)
		err = p.server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		cherryLogger.Infof("run error = %s", err)
	}
}

func (p *HttpServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	for _, controller := range p.controllers {
		controller.Stop()
	}

	if err := p.server.Shutdown(ctx); err != nil {
		cherryLogger.Info(err.Error())
	}

	cherryLogger.Infof("shutdown http server on %s", p.server.Addr)
}
