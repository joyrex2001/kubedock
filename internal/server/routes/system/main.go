package system

import (
	"github.com/gin-gonic/gin"
)

// systemRouter is the object that facilitate all system API endpoints.
type systemRouter struct {
}

// New will instantiate a systemRouter object.
func New(router *gin.Engine) *systemRouter {
	sr := &systemRouter{}
	sr.initRoutes(router)
	return sr
}

// initRoutes will add all suported routes.
func (sr *systemRouter) initRoutes(router *gin.Engine) {
	router.GET("/info", sr.Info)
	router.GET("/version", sr.Version)
	router.GET("/healthz", sr.Healthz)
}
