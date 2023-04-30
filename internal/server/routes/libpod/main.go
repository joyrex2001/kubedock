package libpod

import (
	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes"
)

// RegisterRoutes will add all suported routes.
func RegisterRoutes(router *gin.Engine, cr *routes.ContextRouter) {
	wrap := func(fn func(*routes.ContextRouter, *gin.Context)) gin.HandlerFunc {
		return func(c *gin.Context) {
			fn(cr, c)
		}
	}

	router.GET("/libpod/_ping", wrap(Ping))

	router.POST("/libpod/images/pull", wrap(ImagePull))
	router.GET("/libpod/images/json", wrap(ImageList))

	router.DELETE("/libpod/containers/:id", wrap(ContainerDelete))
	router.POST("/libpod/containers/:id/start", wrap(ContainerStart))
	router.POST("/libpod/containers/:id/stop", wrap(ContainerStop))
	router.POST("/libpod/containers/:id/restart", wrap(ContainerRestart))
	router.POST("/libpod/containers/:id/wait", wrap(ContainerWait))
	router.GET("/libpod/containers/:id/logs", wrap(ContainerLogs))
	router.POST("/libpod/containers/:id/kill", wrap(ContainerKill))
	router.POST("/libpod/containers/:id/rename", wrap(ContainerRename))

	// TODO: make compatible
	router.POST("/libpod/containers/create", wrap(ContainerCreate))

	// not supported podman api at the moment
	router.GET("/libpod/info", httputil.NotImplemented)
	router.POST("/libpod/images/build", httputil.NotImplemented)
}