package container

import (
	"github.com/gin-gonic/gin"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// Container is the object that facilitate all container related
// API endpoints.
type Container struct {
}

// New will instantiate a Container object.
func New(router *gin.Engine) *Container {
	cn := &Container{}
	cn.initRoutes(router)
	return cn
}

// initRoutes will add all suported routes.
func (cn *Container) initRoutes(router *gin.Engine) {
	router.POST("/containers/create", cn.ContainerCreate)
	router.POST("/containers/:id/start", cn.ContainerStart)
	router.GET("/containers/:id/logs", httputil.NotImplemented)
	router.GET("/containers/:id/json", cn.ContainerInfo)
	router.POST("/containers/:id/stop", httputil.NotImplemented)
	router.POST("/containers/:id/kill", httputil.NotImplemented)
	router.DELETE("/containers/:id", cn.ContainerDelete)
	router.POST("/containers/:id/exec", cn.ContainerExec)
	router.POST("/exec/:id/start", cn.ExecStart)
	router.GET("/exec/:id/json", cn.ExecInfo)

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
