package cherryGin

import (
	cherryCode "github.com/cherry-game/cherry/code"
	"github.com/cherry-game/cherry/extend/string"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

const (
	contentType     = "Content-Type"
	htmlContentType = "text/html; charset=utf-8"
	jsonContentType = "application/json; charset=utf-8"
)

type Context struct {
	*gin.Context
}

func (g *Context) GetBody() string {
	data, err := ioutil.ReadAll(g.Request.Body)
	if err != nil {
		return string(data)
	}
	return ""
}

func (g *Context) GetParams(checkPost ...bool) map[string]string {
	maps := make(map[string]string)

	q := g.Context.Request.URL.Query()
	for s, strings := range q {
		maps[s] = strings[0]
	}

	for _, param := range g.Params {
		maps[param.Key] = param.Value
	}

	if len(checkPost) > 0 && checkPost[0] == true {
		g.Request.ParseForm()
		for k, v := range g.Request.PostForm {
			maps[k] = v[0]
		}
	}
	return maps
}

func (g *Context) GetBool(name string, defaultValue bool, checkPost ...bool) bool {
	value := g.GetString(name, "", checkPost...)
	if value == "" {
		return defaultValue
	}

	intValue, ok := cherryString.ToInt(value)
	if ok {
		return intValue > 0
	}

	return defaultValue
}

func (g *Context) GetInt(name string, defaultValue int, checkPost ...bool) int {
	value := g.GetString(name, "", checkPost...)
	if value == "" {
		return defaultValue
	}

	intValue, ok := cherryString.ToInt(value)
	if ok {
		return intValue
	}

	return defaultValue
}

func (g *Context) GetInt32(name string, defaultValue int32, checkPost ...bool) int32 {
	value := g.GetString(name, "", checkPost...)
	if value == "" {
		return defaultValue
	}

	intValue, ok := cherryString.ToInt32(value)
	if ok {
		return intValue
	}

	return defaultValue
}

func (g *Context) GetInt64(name string, defaultValue int64, checkPost ...bool) int64 {
	value := g.GetString(name, "", checkPost...)
	if value == "" {
		return defaultValue
	}

	intValue, ok := cherryString.ToInt64(value)
	if ok {
		return intValue
	}

	return defaultValue
}

func (g *Context) GetString(name, defaultValue string, checkPost ...bool) string {
	if value := g.Param(name); value != "" {
		return value
	}

	if value, ok := g.GetQuery(name); ok {
		return value
	}

	if len(checkPost) > 0 && checkPost[0] == true {
		return g.PostString(name, defaultValue)
	}
	return defaultValue
}

func (g *Context) PostInt(name string, defaultValue int) int {
	if value, ok := g.GetPostForm(name); ok {
		if v, k := cherryString.ToInt(value); k {
			return v
		}
	}
	return defaultValue
}

func (g *Context) PostInt64(name string, defaultValue int64) int64 {
	if value, ok := g.GetPostForm(name); ok {
		if v, k := cherryString.ToInt64(value); k {
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
	g.Header(contentType, htmlContentType)
	g.String(http.StatusOK, html)
}

func (g *Context) RenderJsonString(json string) {
	g.Header(contentType, jsonContentType)
	g.String(http.StatusOK, json)
}

func (g *Context) RenderDataResult(code int32, data ...interface{}) {
	result := cherryCode.NewDataResult(code)
	if len(data) > 0 {
		result.Data = data[0]
	}

	g.Context.JSON(http.StatusOK, &result)
}
