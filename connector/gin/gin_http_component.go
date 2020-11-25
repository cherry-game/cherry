package cherryGin

import (
	"context"
	"github.com/cherry-game/cherry/interfaces"
	"github.com/cherry-game/cherry/logger"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type GinHttpComponentOptions struct {
	name              string        // component name
	ReadTimeout       time.Duration // http server parameter
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

// GinHttpComponent wrapper gin
type GinHttpComponent struct {
	cherryInterfaces.BaseComponent
	*gin.Engine
	server   *http.Server
	name     string
	addr     string
	certFile string
	keyFile  string
}

func New(name string) *GinHttpComponent {
	component := NewWithOptions(GinHttpComponentOptions{
		name:         name,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	logger := cherryLogger.Logger()

	//add default middleware
	component.Use(
		GinDefaultZap(logger),
		RecoveryWithZap(logger, false),
	)

	return component
}

func NewWithOptions(options GinHttpComponentOptions) *GinHttpComponent {
	return &GinHttpComponent{
		Engine: gin.New(),
		name:   options.name,
		server: &http.Server{
			ReadTimeout:       options.ReadTimeout,
			ReadHeaderTimeout: options.ReadHeaderTimeout,
			WriteTimeout:      options.WriteTimeout,
			IdleTimeout:       options.IdleTimeout,
			MaxHeaderBytes:    options.MaxHeaderBytes,
		},
	}
}

func (g *GinHttpComponent) Run(addr string) {
	g.addr = addr
}

// Name unique components name
func (g *GinHttpComponent) Name() string {
	return g.name
}

func (g *GinHttpComponent) Init() {
	if g.addr == "" {
		cherryLogger.Infof("[%s] no set addr value.", g.name)
		return
	}

	g.server.Addr = g.addr
	g.server.Handler = g.Engine

	go func() {
		var err error

		if g.certFile != "" && g.keyFile != "" {
			cherryLogger.Infof("[%s] -> https init. address = %s, certFile = %s, keyFile = %s",
				g.name, g.addr, g.certFile, g.keyFile)
			err = g.server.ListenAndServeTLS(g.certFile, g.keyFile)
		} else {
			cherryLogger.Infof("[%s] -> http init. address = %s", g.name, g.addr)
			err = g.server.ListenAndServe()
		}

		if err != nil {
			cherryLogger.Infof("[%s] run result = %s", g.name, err)
		}
	}()
}

func (g *GinHttpComponent) AfterInit() {

}

func (g *GinHttpComponent) BeforeStop() {

}

func (g *GinHttpComponent) Stop() {
	err := g.server.Shutdown(context.Background())
	cherryLogger.Infof("[%s] shutdown gin http component on %s", g.name, g.addr)

	if err != nil {
		cherryLogger.Info(err.Error())
	}
}

func (g *GinHttpComponent) RunTLS(addr, certFile, keyFile string) {
	g.addr = addr
	g.certFile = certFile
	g.keyFile = keyFile
}

func (g *GinHttpComponent) RunUnix(file string) {
	cherryLogger.Panicf("[%s] not implemented. file = %s", g.name, file)
}

func (g *GinHttpComponent) RunFd(fd int) {
	cherryLogger.Panicf("[%s] not implemented. fd = %d", g.name, fd)
}

func (g *GinHttpComponent) RunListener(listener net.Listener) {
	cherryLogger.Panicf("[%s] not implemented. listener = %s", g.name, listener)
}
