package cherryGin

import "github.com/gin-gonic/gin"

type Group struct {
	*gin.RouterGroup
}

func (p *Group) Any(relativePath string, handlers ...HandlerFunc) {
	p.RouterGroup.Any(relativePath, BindHandlers(handlers)...)
}

func (p *Group) GET(relativePath string, handlers ...HandlerFunc) {
	p.RouterGroup.GET(relativePath, BindHandlers(handlers)...)
}

func (p *Group) POST(relativePath string, handlers ...HandlerFunc) {
	p.RouterGroup.POST(relativePath, BindHandlers(handlers)...)
}
