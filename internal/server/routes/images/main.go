package images

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// imagesRouter is the object that facilitate all image API endpoints.
type imagesRouter struct {
}

// New will instantiate a imagesRouter object.
func New(version int, router *gin.Engine) *imagesRouter {
	vprefix := ""
	if version != 0 {
		vprefix = fmt.Sprintf("/v1.%d", version)
	}
	ir := &imagesRouter{}
	ir.initRoutes(vprefix, router)
	return ir
}

// initRoutes will add all suported routes.
func (ir *imagesRouter) initRoutes(version string, router *gin.Engine) {
	router.GET(version+"/images/json", ir.ImageList)
	router.POST(version+"/images/create", ir.ImageCreate)
	router.GET(version+"/images/:image/*json", ir.ImageJSON)
}
