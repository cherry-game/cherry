package cherryGin

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Context struct {
	*gin.Context
}

func (g *Context) GetInt(name string, defaultValue int) int {
	if value, ok := g.GetQuery(name); ok {
		if v, e := strconv.Atoi(value); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *Context) GetInt64(name string, defaultValue int64) int64 {
	if value, ok := g.GetQuery(name); ok {
		if v, e := strconv.ParseInt(value, 10, 64); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *Context) GetString(name string, defaultValue string) string {
	if value, ok := g.GetQuery(name); ok {
		return value
	}
	return defaultValue
}

func (g *Context) PostInt(name string, defaultValue int) int {
	if value, ok := g.GetPostForm(name); ok {
		if v, e := strconv.Atoi(value); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *Context) PostInt64(name string, defaultValue int64) int64 {
	if value, ok := g.GetPostForm(name); ok {
		if v, e := strconv.ParseInt(value, 10, 64); e == nil {
			return v
		}
	}
	return defaultValue
}

func (g *Context) PostString(name string, defaultValue string) string {
	if value, ok := g.GetPostForm(name); ok {
		return value
	}
	return defaultValue
}

func (g *Context) RenderJSON(value interface{}) {
	g.Context.JSON(http.StatusOK, value)
}

func (g *Context) RenderHTML(html string) {
	g.Header("Content-Type", "text/html; charset=utf-8")
	g.String(http.StatusOK, html)
}

func (g *Context) RenderError(code int, msg string) {
	g.Context.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
	})
}
