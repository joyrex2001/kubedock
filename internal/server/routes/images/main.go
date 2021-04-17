package images

import (
	"github.com/gin-gonic/gin"
)

// imagesRouter is the object that facilitate all image API endpoints.
type imagesRouter struct {
}

// New will instantiate a imagesRouter object.
func New(router *gin.Engine) *imagesRouter {
	ir := &imagesRouter{}
	ir.initRoutes(router)
	return ir
}

// initRoutes will add all suported routes.
func (ir *imagesRouter) initRoutes(router *gin.Engine) {
	router.GET("/images/json", ir.ImageList)
	router.POST("/images/create", ir.ImageCreate)
	router.GET("/images/:image/*json", ir.ImageJSON)
}
