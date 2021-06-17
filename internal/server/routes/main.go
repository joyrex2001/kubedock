package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/model"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

// Config is the structure to instantiate a Router object
type Config struct {
	// Inspector specifies if the image inspect feature is enabled
	Inspector bool
	// PortForward specifies if the the services should be port-forwarded
	PortForward bool
	// ReverseProxy enables a reverse-proxy to the services on localhost
	ReverseProxy bool
	// RequestCPU contains an optional default k8s cpu request
	RequestCPU string
	// RequestMemory contains an optional default k8s memory request
	RequestMemory string
}

// Router is the object that facilitates the kubedock API endpoints.
type Router struct {
	db  *model.Database
	kub backend.Backend
	cfg Config
}

// New will instantiate a containerRouter object.
func New(router *gin.Engine, kub backend.Backend, cfg Config) (*Router, error) {
	db, err := model.New()
	if err != nil {
		return nil, err
	}
	cr := &Router{
		db:  db,
		kub: kub,
		cfg: cfg,
	}
	cr.initRoutes(router)
	return cr, nil
}

// initRoutes will add all suported routes.
func (cr *Router) initRoutes(router *gin.Engine) {
	router.GET("/info", cr.Info)
	router.GET("/version", cr.Version)
	router.GET("/_ping", cr.Ping)
	router.HEAD("/_ping", cr.Ping)

	router.POST("/containers/create", cr.ContainerCreate)
	router.POST("/containers/:id/start", cr.ContainerStart)
	router.POST("/containers/:id/attach", cr.ContainerAttach)
	router.POST("/containers/:id/exec", cr.ContainerExec)
	router.POST("/containers/:id/stop", cr.ContainerStop)
	router.POST("/containers/:id/kill", cr.ContainerKill)
	router.POST("/containers/:id/wait", cr.ContainerWait)
	router.DELETE("/containers/:id", cr.ContainerDelete)
	router.GET("/containers/json", cr.ContainerList)
	router.GET("/containers/:id/json", cr.ContainerInfo)
	router.GET("/containers/:id/logs", cr.ContainerLogs)
	router.PUT("/containers/:id/archive", cr.PutArchive)

	router.POST("/exec/:id/start", cr.ExecStart)
	router.GET("/exec/:id/json", cr.ExecInfo)

	router.POST("/networks/create", cr.NetworksCreate)
	router.POST("/networks/:id/connect", cr.NetworksConnect)
	router.POST("/networks/:id/disconnect", cr.NetworksDisconnect)
	router.GET("/networks", cr.NetworksList)
	router.GET("/networks/:id", cr.NetworksInfo)
	router.DELETE("/networks/:id", cr.NetworksDelete)

	router.POST("/images/create", cr.ImageCreate)
	router.GET("/images/json", cr.ImageList)
	router.GET("/images/:image/*json", cr.ImageJSON)

	// not supported at the moment
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
	router.GET("/containers/:id/attach/ws", httputil.NotImplemented)
	router.HEAD("/containers/:id/archive", httputil.NotImplemented)
	router.GET("/containers/:id/archive", httputil.NotImplemented)
	router.POST("/containers/prune", httputil.NotImplemented)
	router.GET("/networks/reaper_default", httputil.NotImplemented)
	router.POST("/build", httputil.NotImplemented)
}
