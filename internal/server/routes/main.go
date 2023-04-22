package routes

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/model"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

const (
	// PollRate defines maximum polling request per second towards the backend
	PollRate = 1
	// PollBurst defines maximum burst poll requests towards the backend
	PollBurst = 3
)

// Config is the structure to instantiate a Router object
type Config struct {
	// Inspector specifies if the image inspect feature is enabled
	Inspector bool
	// PortForward specifies if the the services should be port-forwarded
	PortForward bool
	// ReverseProxy enables a reverse-proxy to the services via 0.0.0.0 on the kubedock host
	ReverseProxy bool
	// RequestCPU contains an optional default k8s cpu request
	RequestCPU string
	// RequestMemory contains an optional default k8s memory request
	RequestMemory string
	// RunasUser contains the UID to run pods as
	RunasUser string
	// PullPolicy contains the default pull policy for images
	PullPolicy string
	// PreArchive will enable copying files without starting containers
	PreArchive bool
	// DeployAsJob will deploy containers as jobs instead of deployments
	DeployAsJob bool
	// ServiceAccount contains the service account name to be used for running containers
	ServiceAccount string
}

// Router is the object that facilitates the kubedock API endpoints.
type Router struct {
	db     *model.Database
	kub    backend.Backend
	plim   *rate.Limiter
	cfg    Config
	events events.Events
}

// New will instantiate a containerRouter object.
func New(router *gin.Engine, kub backend.Backend, cfg Config) (*Router, error) {
	db, err := model.New()
	if err != nil {
		return nil, err
	}
	cr := &Router{
		db:     db,
		kub:    kub,
		plim:   rate.NewLimiter(PollRate, PollBurst),
		cfg:    cfg,
		events: events.New(),
	}
	cr.initRoutes(router)
	return cr, nil
}

// initRoutes will add all suported routes.
func (cr *Router) initRoutes(router *gin.Engine) {
	router.GET("/info", cr.Info)
	router.GET("/events", cr.Events)
	router.GET("/version", cr.Version)
	router.GET("/_ping", cr.Ping)
	router.HEAD("/_ping", cr.Ping)

	router.POST("/containers/create", cr.ContainerCreate)
	router.POST("/containers/:id/start", cr.ContainerStart)
	router.POST("/containers/:id/attach", cr.ContainerAttach)
	router.POST("/containers/:id/exec", cr.ContainerExec)
	router.POST("/containers/:id/stop", cr.ContainerStop)
	router.POST("/containers/:id/restart", cr.ContainerRestart)
	router.POST("/containers/:id/kill", cr.ContainerKill)
	router.POST("/containers/:id/wait", cr.ContainerWait)
	router.POST("/containers/:id/resize", cr.ContainerResize)
	router.DELETE("/containers/:id", cr.ContainerDelete)
	router.GET("/containers/json", cr.ContainerList)
	router.GET("/containers/:id/json", cr.ContainerInfo)
	router.GET("/containers/:id/logs", cr.ContainerLogs)
	router.GET("/containers/:id/archive", cr.GetArchive)
	router.PUT("/containers/:id/archive", cr.PutArchive)
	router.POST("/containers/:id/rename", cr.ContainerRename)

	router.POST("/exec/:id/start", cr.ExecStart)
	router.GET("/exec/:id/json", cr.ExecInfo)

	router.POST("/networks/create", cr.NetworksCreate)
	router.POST("/networks/:id/connect", cr.NetworksConnect)
	router.POST("/networks/:id/disconnect", cr.NetworksDisconnect)
	router.GET("/networks", cr.NetworksList)
	router.GET("/networks/:id", cr.NetworksInfo)
	router.DELETE("/networks/:id", cr.NetworksDelete)
	router.POST("/networks/prune", cr.NetworksPrune)

	router.POST("/images/create", cr.ImageCreate)
	router.GET("/images/json", cr.ImageList)
	router.GET("/images/:image/*json", cr.ImageJSON)

	// not supported at the moment
	router.GET("/containers/:id/top", httputil.NotImplemented)
	router.GET("/containers/:id/changes", httputil.NotImplemented)
	router.GET("/containers/:id/export", httputil.NotImplemented)
	router.GET("/containers/:id/stats", httputil.NotImplemented)
	router.POST("/containers/:id/update", httputil.NotImplemented)
	router.POST("/containers/:id/pause", httputil.NotImplemented)
	router.POST("/containers/:id/unpause", httputil.NotImplemented)
	router.GET("/containers/:id/attach/ws", httputil.NotImplemented)
	router.HEAD("/containers/:id/archive", httputil.NotImplemented)
	router.POST("/containers/prune", httputil.NotImplemented)
	router.POST("/build", httputil.NotImplemented)
	router.GET("/volumes", httputil.NotImplemented)
	router.GET("/volumes/:id", httputil.NotImplemented)
	router.DELETE("/volumes/:id", httputil.NotImplemented)
	router.POST("/volumes/create", httputil.NotImplemented)
	router.POST("/volumes/prune", httputil.NotImplemented)
}
