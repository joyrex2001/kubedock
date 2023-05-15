package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
	"github.com/joyrex2001/kubedock/internal/server/routes/docker"
)

// RegisterDockerRoutes will add all suported docker routes.
func RegisterDockerRoutes(router *gin.Engine, cr *common.ContextRouter) {
	wrap := func(fn func(*common.ContextRouter, *gin.Context)) gin.HandlerFunc {
		return func(c *gin.Context) {
			fn(cr, c)
		}
	}

	router.GET("/info", wrap(docker.Info))
	router.GET("/events", wrap(docker.Events))
	router.GET("/version", wrap(docker.Version))
	router.GET("/_ping", wrap(docker.Ping))
	router.HEAD("/_ping", wrap(docker.Ping))

	router.POST("/containers/create", wrap(docker.ContainerCreate))
	router.POST("/containers/:id/start", wrap(common.ContainerStart))
	router.POST("/containers/:id/attach", wrap(common.ContainerAttach))
	router.POST("/containers/:id/exec", wrap(docker.ContainerExec))
	router.POST("/containers/:id/stop", wrap(common.ContainerStop))
	router.POST("/containers/:id/restart", wrap(common.ContainerRestart))
	router.POST("/containers/:id/kill", wrap(common.ContainerKill))
	router.POST("/containers/:id/wait", wrap(docker.ContainerWait))
	router.POST("/containers/:id/resize", wrap(common.ContainerResize))
	router.DELETE("/containers/:id", wrap(common.ContainerDelete))
	router.GET("/containers/json", wrap(docker.ContainerList))
	router.GET("/containers/:id/json", wrap(docker.ContainerInfo))
	router.GET("/containers/:id/logs", wrap(common.ContainerLogs))
	router.HEAD("/containers/:id/archive", wrap(docker.HeadArchive))
	router.GET("/containers/:id/archive", wrap(docker.GetArchive))
	router.PUT("/containers/:id/archive", wrap(docker.PutArchive))
	router.POST("/containers/:id/rename", wrap(docker.ContainerRename))

	router.POST("/exec/:id/start", wrap(docker.ExecStart))
	router.GET("/exec/:id/json", wrap(docker.ExecInfo))

	router.POST("/networks/create", wrap(docker.NetworksCreate))
	router.POST("/networks/:id/connect", wrap(docker.NetworksConnect))
	router.POST("/networks/:id/disconnect", wrap(docker.NetworksDisconnect))
	router.GET("/networks", wrap(docker.NetworksList))
	router.GET("/networks/:id", wrap(docker.NetworksInfo))
	router.DELETE("/networks/:id", wrap(docker.NetworksDelete))
	router.POST("/networks/prune", wrap(docker.NetworksPrune))

	router.POST("/images/create", wrap(docker.ImageCreate))
	router.GET("/images/json", wrap(common.ImageList))
	router.GET("/images/:image/*json", wrap(docker.ImageJSON))

	// not supported docker api at the moment
	router.GET("/containers/:id/top", httputil.NotImplemented)
	router.GET("/containers/:id/changes", httputil.NotImplemented)
	router.GET("/containers/:id/export", httputil.NotImplemented)
	router.GET("/containers/:id/stats", httputil.NotImplemented)
	router.POST("/containers/:id/update", httputil.NotImplemented)
	router.POST("/containers/:id/pause", httputil.NotImplemented)
	router.POST("/containers/:id/unpause", httputil.NotImplemented)
	router.GET("/containers/:id/attach/ws", httputil.NotImplemented)
	router.POST("/containers/prune", httputil.NotImplemented)
	router.POST("/build", httputil.NotImplemented)
	router.GET("/volumes", httputil.NotImplemented)
	router.GET("/volumes/:id", httputil.NotImplemented)
	router.DELETE("/volumes/:id", httputil.NotImplemented)
	router.POST("/volumes/create", httputil.NotImplemented)
	router.POST("/volumes/prune", httputil.NotImplemented)
}
