package networks

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET "/networks"
func (nr *networksRouter) NetworksList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// POST "/networks/create"
func (nr *networksRouter) NetworksCreate(c *gin.Context) {
	// from := c.Query("fromImage")
	c.JSON(http.StatusOK, gin.H{
		"status": "Download complete",
		// TODO: add progressdetail...
	})
}
