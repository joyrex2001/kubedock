package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET "/networks"
func (nr *Router) NetworksList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// POST "/networks/create"
func (nr *Router) NetworksCreate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "sure, fine. Party time. Excellent.",
	})
}
