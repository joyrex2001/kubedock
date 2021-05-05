package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImageList - list Images. Stubbed, not relevant on k8s.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageList
// GET "/images/json"
func (cr *Router) ImageList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// ImageJSON - return low-level information about an image.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageInspect
// GET "/images/:image/json"
func (cr *Router) ImageJSON(c *gin.Context) {
	id := c.Param("image")
	c.JSON(http.StatusOK, gin.H{
		"Id":      id,
		"Created": "2018-12-18T01:20:53.669016181Z",
		"Size":    0,
		"ContainerConfig": gin.H{
			"Image": id,
		},
	})
}

// ImageCreate - create an image.
// https://docs.docker.com/engine/api/v1.41/#operation/ImageCreate
// POST "/images/create"
func (cr *Router) ImageCreate(c *gin.Context) {
	// from := c.Query("fromImage")
	c.JSON(http.StatusOK, gin.H{
		"status": "Download complete",
		// TODO: add progressdetail...
	})
}
