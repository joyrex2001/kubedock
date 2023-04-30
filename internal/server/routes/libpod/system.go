package libpod

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping - dummy endpoint you can use to test if the server is accessible.
// https://docs.docker.com/engine/api/v1.41/#operation/SystemPing
// HEAD "/_ping"
// GET "/_ping"
func (cr *Router) Ping(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
