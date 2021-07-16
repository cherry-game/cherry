package main

import (
	"github.com/cherry-game/cherry/component/gin"
	cherryResult "github.com/cherry-game/cherry/extend/result"
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
	t.GET("/render_result", t.renderResult)

	cherrySnowflake.SetDefaultNode(1)

	loadResult()
}

func (t *Test1Controller) index(c *cherryGin.Context) {
	c.RenderHTML("this is index... " + cherrySnowflake.Next().String())
}

func (t *Test1Controller) panic(c *gin.Context) {
	c.String(http.StatusOK, "test panic")
	panic("test panic!")
}

func loadResult() {
	var resultList []*cherryResult.Result

	result1 := cherryResult.New(0)
	result1.Message = "成功"
	resultList = append(resultList, result1)

	cherryGin.InitResult(resultList...)
}

func (t *Test1Controller) renderResult(c *cherryGin.Context) {
	str := cherrySnowflake.Next().Base58()
	c.RenderResult(0, str)
}
