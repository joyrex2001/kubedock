package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NetworksList - list networks.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkList
// GET "/networks"
func (nr *Router) NetworksList(c *gin.Context) {
	c.JSON(http.StatusOK, []string{})
}

// NetworksCreate - create a network.
// https://docs.docker.com/engine/api/v1.41/#operation/NetworkCreate
// POST "/networks/create"
func (nr *Router) NetworksCreate(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"Id":      "dummy-network-id",
		"Warning": "",
	})
}
