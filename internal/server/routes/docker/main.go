package docker

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

	router.GET("/info", wrap(Info))
	router.GET("/events", wrap(Events))
	router.GET("/version", wrap(Version))
	router.GET("/_ping", wrap(Ping))
	router.HEAD("/_ping", wrap(Ping))

	router.POST("/containers/create", wrap(ContainerCreate))
	router.POST("/containers/:id/start", wrap(ContainerStart))
	router.POST("/containers/:id/attach", wrap(ContainerAttach))
	router.POST("/containers/:id/exec", wrap(ContainerExec))
	router.POST("/containers/:id/stop", wrap(ContainerStop))
	router.POST("/containers/:id/restart", wrap(ContainerRestart))
	router.POST("/containers/:id/kill", wrap(ContainerKill))
	router.POST("/containers/:id/wait", wrap(ContainerWait))
	router.POST("/containers/:id/resize", wrap(ContainerResize))
	router.DELETE("/containers/:id", wrap(ContainerDelete))
	router.GET("/containers/json", wrap(ContainerList))
	router.GET("/containers/:id/json", wrap(ContainerInfo))
	router.GET("/containers/:id/logs", wrap(ContainerLogs))
	router.HEAD("/containers/:id/archive", wrap(HeadArchive))
	router.GET("/containers/:id/archive", wrap(GetArchive))
	router.PUT("/containers/:id/archive", wrap(PutArchive))
	router.POST("/containers/:id/rename", wrap(ContainerRename))

	router.POST("/exec/:id/start", wrap(ExecStart))
	router.GET("/exec/:id/json", wrap(ExecInfo))

	router.POST("/networks/create", wrap(NetworksCreate))
	router.POST("/networks/:id/connect", wrap(NetworksConnect))
	router.POST("/networks/:id/disconnect", wrap(NetworksDisconnect))
	router.GET("/networks", wrap(NetworksList))
	router.GET("/networks/:id", wrap(NetworksInfo))
	router.DELETE("/networks/:id", wrap(NetworksDelete))
	router.POST("/networks/prune", wrap(NetworksPrune))

	router.POST("/images/create", wrap(ImageCreate))
	router.GET("/images/json", wrap(ImageList))
	router.GET("/images/:image/*json", wrap(ImageJSON))

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
