package cherryGin

import (
	"io"
	"net/http"

	cslice "github.com/cherry-game/cherry/extend/slice"
	cstring "github.com/cherry-game/cherry/extend/string"
	"github.com/gin-gonic/gin"
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
	data, err := io.ReadAll(g.Request.Body)
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

	if len(checkPost) > 0 && checkPost[0] {
		err := g.Request.ParseForm()
		if err != nil {
			return maps
		}

		for k, v := range g.Request.PostForm {
			maps[k] = v[0]
		}
	}
	return maps
}

func (g *Context) IsPost() bool {
	return g.Context.Request.Method == http.MethodPost
}

func (g *Context) IsGet() bool {
	return g.Context.Request.Method == http.MethodGet
}

func (g *Context) GetBool(name string, defaultValue bool, checkPost ...bool) bool {
	value := g.GetString(name, "", checkPost...)
	if value == "" {
		return defaultValue
	}

	intValue, ok := cstring.ToInt(value)
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

	intValue, ok := cstring.ToInt(value)
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

	intValue, ok := cstring.ToInt32(value)
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

	intValue, _ := cstring.ToInt64(value, defaultValue)
	return intValue
}

func (g *Context) GetString(name, defaultValue string, checkPost ...bool) string {
	if value := g.Param(name); value != "" {
		return value
	}

	if value, ok := g.GetQuery(name); ok {
		return value
	}

	if len(checkPost) > 0 && checkPost[0] {
		return g.PostString(name, defaultValue)
	}
	return defaultValue
}

func (g *Context) PostInt(name string, defaultValue int) int {
	if value, ok := g.GetPostForm(name); ok {
		if v, k := cstring.ToInt(value); k {
			return v
		}
	}
	return defaultValue
}

func (g *Context) PostInt32(name string, defaultValue int32) int32 {
	if value, ok := g.GetPostForm(name); ok {
		if v, k := cstring.ToInt32(value); k {
			return v
		}
	}
	return defaultValue
}

func (g *Context) PostInt64(name string, defaultValue int64) int64 {
	if value, ok := g.GetPostForm(name); ok {
		if v, k := cstring.ToInt64(value); k {
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

func (g *Context) PostFormIntArray(name string) []int {
	array := g.PostFormArray(name)
	if len(array) < 1 {
		return []int{}
	}

	return cslice.StringToInt(array)
}

func (g *Context) PostFormInt32Array(name string) []int32 {
	array := g.PostFormArray(name)
	if len(array) < 1 {
		return []int32{}
	}

	return cslice.StringToInt32(array)
}

func (g *Context) PostFormInt64Array(name string) []int64 {
	array := g.PostFormArray(name)
	if len(array) < 1 {
		return []int64{}
	}

	return cslice.StringToInt64(array)
}

func (g *Context) HTML200(name string, obj ...interface{}) {
	if len(obj) > 0 {
		g.HTML(http.StatusOK, name, obj[0])
	} else {
		g.HTML(http.StatusOK, name, nil)
	}
}

func (g *Context) JSON200(obj interface{}) {
	g.JSON(http.StatusOK, obj)
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

func (g *Context) GetIntCookie(name string, defaultValue int) int {
	value, err := g.Cookie(name)
	if err != nil {
		return defaultValue
	}

	if v, k := cstring.ToInt(value); k {
		return v
	}
	return defaultValue
}

func (g *Context) GetInt32Cookie(name string, defaultValue int32) int32 {
	value, err := g.Cookie(name)
	if err != nil {
		return defaultValue
	}

	if v, k := cstring.ToInt32(value); k {
		return v
	}
	return defaultValue
}

func (g *Context) GetInt64Cookie(name string, defaultValue int64) int64 {
	value, err := g.Cookie(name)
	if err != nil {
		return defaultValue
	}

	if v, k := cstring.ToInt64(value); k {
		return v
	}
	return defaultValue
}

func (g *Context) GetStringCookie(name string, defaultValue string) string {
	value, err := g.Cookie(name)
	if err != nil {
		return defaultValue
	}

	return value
}
