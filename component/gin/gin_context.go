package cherryGin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type GinContext struct {
	*gin.Context
}

func (g *GinContext) GetInt(name string, defaultValue int) int {
	if value, ok := g.GetQuery(name); ok {
		if v, e := strconv.Atoi(value); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *GinContext) GetInt64(name string, defaultValue int64) int64 {
	if value, ok := g.GetQuery(name); ok {
		if v, e := strconv.ParseInt(value, 10, 64); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *GinContext) GetString(name string, defaultValue string) string {
	if value, ok := g.GetQuery(name); ok {
		return value
	}
	return defaultValue
}

func (g *GinContext) PostInt(name string, defaultValue int) int {
	if value, ok := g.GetPostForm(name); ok {
		if v, e := strconv.Atoi(value); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *GinContext) PostInt64(name string, defaultValue int64) int64 {
	if value, ok := g.GetPostForm(name); ok {
		if v, e := strconv.ParseInt(value, 10, 64); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *GinContext) PostString(name string, defaultValue string) string {
	if value, ok := g.GetPostForm(name); ok {
		return value
	}
	return defaultValue
}

func (g *GinContext) RenderJSON(value interface{}) {
	g.Context.JSON(http.StatusOK, value)
}

func (g *GinContext) RenderHTML(html string) {
	g.String(http.StatusOK, html)
}

func (g *GinContext) RenderError(msg string) {
	g.Context.JSON(http.StatusOK, gin.H{
		"code": 500,
		"msg":  msg,
	})
}
