package container

import (
	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// containerRouter is the object that facilitate all container
// related API endpoints.
type containerRouter struct {
	factory container.Factory
}

// New will instantiate a containerRouter object.
func New(router *gin.Engine, factory container.Factory) *containerRouter {
	cr := &containerRouter{
		factory: factory,
	}
	cr.initRoutes(router)
	return cr
}

// initRoutes will add all suported routes.
func (cr *containerRouter) initRoutes(router *gin.Engine) {
	router.POST("/containers/create", cr.ContainerCreate)
	router.POST("/containers/:id/start", cr.ContainerStart)
	router.GET("/containers/:id/json", cr.ContainerInfo)
	router.DELETE("/containers/:id", cr.ContainerDelete)
	router.POST("/containers/:id/exec", cr.ContainerExec)
	router.POST("/exec/:id/start", cr.ExecStart)
	router.GET("/exec/:id/json", cr.ExecInfo)

	// not supported at the moment
	router.GET("/containers/:id/logs", httputil.NotImplemented)
	router.POST("/containers/:id/stop", httputil.NotImplemented)
	router.POST("/containers/:id/kill", httputil.NotImplemented)
	router.GET("/containers/json", httputil.NotImplemented)
	router.GET("/containers/:id/top", httputil.NotImplemented)
	router.GET("/containers/:id/changes", httputil.NotImplemented)
	router.GET("/containers/:id/export", httputil.NotImplemented)
	router.GET("/containers/:id/stats", httputil.NotImplemented)
	router.POST("/containers/:id/resize", httputil.NotImplemented)
	router.POST("/containers/:id/restart", httputil.NotImplemented)
	router.POST("/containers/:id/update", httputil.NotImplemented)
	router.POST("/containers/:id/rename", httputil.NotImplemented)
	router.POST("/containers/:id/pause", httputil.NotImplemented)
	router.POST("/containers/:id/unpause", httputil.NotImplemented)
	router.POST("/containers/:id/attach", httputil.NotImplemented)
	router.GET("/containers/:id/attach/ws", httputil.NotImplemented)
	router.POST("/containers/:id/wait", httputil.NotImplemented)
	router.HEAD("/containers/:id/archive", httputil.NotImplemented)
	router.GET("/containers/:id/archive", httputil.NotImplemented)
	router.PUT("/containers/:id/archive", httputil.NotImplemented)
	router.POST("/containers/prune", httputil.NotImplemented)
}
