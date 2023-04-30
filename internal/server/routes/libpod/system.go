package libpod

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// Ping - dummy endpoint you can use to test if the server is accessible.
// https://docs.podman.io/en/latest/_static/api.html?version=v4.2#tag/system/operation/SystemPing
// GET "/libpod/_ping"
func Ping(cr *routes.ContextRouter, c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
