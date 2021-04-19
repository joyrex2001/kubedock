package images

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET "/images/json"
func (ir *imagesRouter) ImageList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// GET "/images/:image/json"
func (ir *imagesRouter) ImageJSON(c *gin.Context) {
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

// POST "/images/create"
func (ir *imagesRouter) ImageCreate(c *gin.Context) {
	// from := c.Query("fromImage")
	c.JSON(http.StatusOK, gin.H{
		"status": "Download complete",
		// TODO: add progressdetail...
	})
}
