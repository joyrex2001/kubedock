package routes

import (
	"log"

	"github.com/gin-gonic/gin"
)

// Error will return an error response in json.
func Error(c *gin.Context, status int, err error) {
	log.Print(err)
	c.JSON(status, gin.H{
		"error": err.Error(),
	})
}
