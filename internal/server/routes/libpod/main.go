package libpod

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

	// podman api
	router.GET("/libpod/_ping", cr.Ping)
	router.HEAD("/libpod/_ping", cr.Ping)

	router.POST("/libpod/images/pull", cr.ImagePull)
	router.GET("/libpod/images/json", cr.ImageList)

	router.DELETE("/libpod/containers/:id", cr.ContainerDelete)
	router.POST("/libpod/containers/:id/start", cr.ContainerStart)
	router.POST("/libpod/containers/:id/stop", cr.ContainerStop)
	router.POST("/libpod/containers/:id/restart", cr.ContainerRestart)
	router.POST("/libpod/containers/:id/wait", cr.ContainerWait)
	router.GET("/libpod/containers/:id/logs", cr.ContainerLogs)
	router.POST("/libpod/containers/:id/kill", cr.ContainerKill)
	router.POST("/libpod/containers/:id/rename", cr.ContainerRename)

	// TODO: make compatible
	router.POST("/libpod/containers/create", cr.ContainerCreate)

	// not supported podman api at the moment
	router.GET("/libpod/info", httputil.NotImplemented)
	router.POST("/libpod/images/build", httputil.NotImplemented)
}
