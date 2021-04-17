package system

import (
	"github.com/gin-gonic/gin"
)

// System is the object that facilitate all system API endpoints.
type System struct {
}

// New will instantiate a System object.
func New(router *gin.Engine) *System {
	s := &System{}
	s.initRoutes(router)
	return s
}

// initRoutes will add all suported routes.
func (s *System) initRoutes(router *gin.Engine) {
	router.GET("/info", s.Info)
	router.GET("/version", s.Version)
	router.GET("/healthz", s.Healthz)
}
