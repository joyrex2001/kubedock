package routes

import (
	"golang.org/x/time/rate"

	"github.com/joyrex2001/kubedock/internal/backend"
	"github.com/joyrex2001/kubedock/internal/events"
	"github.com/joyrex2001/kubedock/internal/model"
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

// ContextRouter is the object that contains shared context for the kubedock API endpoints.
type ContextRouter struct {
	Config  Config
	DB      *model.Database
	Backend backend.Backend
	Events  events.Events
	Limiter *rate.Limiter
}

// NewContextRouter will instantiate a ContextRouter object.
func NewContextRouter(kub backend.Backend, cfg Config) (*ContextRouter, error) {
	db, err := model.New()
	if err != nil {
		return nil, err
	}
	cr := &ContextRouter{
		Config:  cfg,
		DB:      db,
		Backend: kub,
		Events:  events.New(),
		Limiter: rate.NewLimiter(PollRate, PollBurst),
	}
	return cr, nil
}
