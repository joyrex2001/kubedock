package system

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// systemRouter is the object that facilitate all system API endpoints.
type systemRouter struct {
}

// New will instantiate a systemRouter object.
func New(version int, router *gin.Engine) *systemRouter {
	vprefix := ""
	if version != 0 {
		vprefix = fmt.Sprintf("/v1.%d", version)
	}
	sr := &systemRouter{}
	sr.initRoutes(vprefix, router)
	return sr
}

// initRoutes will add all suported routes.
func (sr *systemRouter) initRoutes(version string, router *gin.Engine) {
	router.GET(version+"/_ping", sr.Ping)
	router.HEAD(version+"/_ping", sr.Ping)
	router.GET(version+"/info", sr.Info)
	router.GET(version+"/version", sr.Version)
	router.GET(version+"/healthz", sr.Healthz)
}
