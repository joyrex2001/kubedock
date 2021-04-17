package images

import (
	"github.com/gin-gonic/gin"
)

// Images is the object that facilitate all image API endpoints.
type Images struct {
}

// New will instantiate a Images object.
func New(router *gin.Engine) *Images {
	im := &Images{}
	im.initRoutes(router)
	return im
}

// initRoutes will add all suported routes.
func (im *Images) initRoutes(router *gin.Engine) {
	router.GET("/images/json", im.ImageList)
	router.POST("/images/create", im.ImageCreate)
	router.GET("/images/:image/*json", im.ImageJson)
}
