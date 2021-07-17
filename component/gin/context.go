package cherryGin

import (
	"github.com/cherry-game/cherry/extend/result"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
)

var resultMaps = make(map[int]*cherryResult.Result)

func InitResult(list ...*cherryResult.Result) {
	resultMaps = make(map[int]*cherryResult.Result)

	for _, result := range list {
		resultMaps[result.Code] = result
	}
}

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

	if len(checkPost) > 0 && checkPost[0] == true {
		g.Request.ParseForm()

		for k, v := range g.Request.PostForm {
			maps[k] = v[0]
		}
	}
	return maps
}

func (g *Context) GetBool(name string, defaultValue bool, checkPost ...bool) bool {
	var intDefaultValue = 0
	if defaultValue {
		intDefaultValue = 1
	}

	return g.GetInt(name, intDefaultValue, checkPost...) > 0
}

func (g *Context) GetInt(name string, defaultValue int, checkPost ...bool) int {
	if value, ok := g.GetQuery(name); ok {
		if v, e := strconv.Atoi(value); e == nil {
			return v
		}
	}

	if len(checkPost) > 0 && checkPost[0] == true {
		return g.PostInt(name, defaultValue)
	}

	return defaultValue
}

func (g *Context) GetInt64(name string, defaultValue int64, checkPost ...bool) int64 {
	if value, ok := g.GetQuery(name); ok {
		if v, e := strconv.ParseInt(value, 10, 64); e == nil {
			return v
		}
	}

	if len(checkPost) > 0 && checkPost[0] == true {
		return g.PostInt64(name, defaultValue)
	}

	return defaultValue
}

func (g *Context) GetString(name, defaultValue string, checkPost ...bool) string {
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

func (g *Context) RenderResult(code int, data ...interface{}) {
	result, found := resultMaps[code]
	if found == false {
		result = cherryResult.New(code)
	}

	if len(data) > 0 {
		result.Data = data[0]
	}

	g.Context.JSON(http.StatusOK, result)
}
