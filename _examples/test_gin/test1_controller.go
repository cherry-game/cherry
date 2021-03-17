package main

import (
	"github.com/cherry-game/cherry/component/gin"
	"github.com/cherry-game/cherry/extend/snowflake"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Test1Controller struct {
	cherryGin.BaseController
}

func (t *Test1Controller) Init() {
	t.GET("/", t.index)
	t.Engine.GET("/panic", t.panic)

	cherrySnowflake.SetDefaultNode(1)
}

func (t *Test1Controller) index(c *cherryGin.GinContext) {
	c.RenderHTML("this is index... " + cherrySnowflake.Next().String())
}

func (t *Test1Controller) panic(c *gin.Context) {
	c.String(http.StatusOK, "test panic")
	panic("test panic!")
}
