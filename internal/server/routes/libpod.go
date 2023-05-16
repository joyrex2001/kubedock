package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
	"github.com/joyrex2001/kubedock/internal/server/routes/libpod"
)

// RegisterLibpodRoutes will add all suported podman routes.
func RegisterLibpodRoutes(router *gin.Engine, cr *common.ContextRouter) {
	wrap := func(fn func(*common.ContextRouter, *gin.Context)) gin.HandlerFunc {
		return func(c *gin.Context) {
			fn(cr, c)
		}
	}

	router.GET("/libpod/version", wrap(libpod.Version))
	router.GET("/libpod/_ping", wrap(libpod.Ping))
	router.HEAD("/libpod/_ping", wrap(libpod.Ping))

	router.POST("/libpod/images/pull", wrap(libpod.ImagePull))
	router.GET("/libpod/images/json", wrap(common.ImageList))

	router.POST("/libpod/containers/create", wrap(libpod.ContainerCreate))
	router.POST("/libpod/containers/:id/start", wrap(common.ContainerStart))
	router.POST("/libpod/containers/:id/attach", wrap(common.ContainerAttach))
	router.POST("/libpod/containers/:id/stop", wrap(common.ContainerStop))
	router.POST("/libpod/containers/:id/restart", wrap(common.ContainerRestart))
	router.POST("/libpod/containers/:id/kill", wrap(common.ContainerKill))
	router.POST("/libpod/containers/:id/wait", wrap(libpod.ContainerWait))
	router.POST("/libpod/containers/:id/resize", wrap(common.ContainerResize))
	router.DELETE("/libpod/containers/:id", wrap(common.ContainerDelete))
	router.GET("/libpod/containers/json", wrap(libpod.ContainerList))
	router.GET("/libpod/containers/:id/exists", wrap(libpod.ContainerExists))
	router.GET("/libpod/containers/:id/json", wrap(libpod.ContainerInfo))
	router.GET("/libpod/containers/:id/logs", wrap(common.ContainerLogs))
	router.POST("/libpod/containers/:id/rename", wrap(common.ContainerRename))

	router.POST("/libpod/containers/:id/exec", wrap(common.ContainerExec))
	router.POST("/libpod/exec/:id/start", wrap(common.ExecStart))
	router.GET("/libpod/exec/:id/json", wrap(common.ExecInfo))

	// not supported podman api at the moment
	router.GET("/libpod/info", httputil.NotImplemented)
	router.POST("/libpod/images/build", httputil.NotImplemented)
}
