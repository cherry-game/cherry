package main

import (
	cherryGin "github.com/cherry-game/cherry/component/gin"
	cherrySnowflake "github.com/cherry-game/cherry/extend/snowflake"
	cherryInterfaces "github.com/cherry-game/cherry/interfaces"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Test1Controller struct {
	cherryGin.BaseIController
	app cherryInterfaces.IApplication
}

func (t *Test1Controller) Init(app cherryInterfaces.IApplication, engine *gin.Engine) {
	t.app = app
	engine.GET("/", t.index)
	engine.GET("/panic", t.panic)

	cherrySnowflake.SetDefaultNode(1)
}

func (t *Test1Controller) index(c *gin.Context) {
	c.String(http.StatusOK, "this is index... "+cherrySnowflake.Next().String())
}

func (t *Test1Controller) panic(c *gin.Context) {
	c.String(http.StatusOK, "test panic")
	panic("test panic!")
}
