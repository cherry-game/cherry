package cherryGin

import "github.com/gin-gonic/gin"

type GinContext struct {
	GinHttpComponent
	*gin.Context
}
