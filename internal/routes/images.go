package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET "/images/json"
func ImageList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// POST "/images/create"
func ImageCreate(c *gin.Context) {
	// from := c.Query("fromImage")
	c.JSON(http.StatusOK, gin.H{
		"status": "Download complete",
		// TODO: add progressdetail...
	})
}
