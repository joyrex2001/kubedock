package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error will return an error response in json.
func Error(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err,
	})
}
