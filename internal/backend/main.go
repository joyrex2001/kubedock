package backend

import (
	"io"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// Backend is the interface to orchestrate and manage kubernetes objects.
type Backend interface {
	StartContainer(*types.Container) error
	DeleteContainer(*types.Container) error
	DeleteContainersOlderThan(time.Duration) error
	CopyToContainer(*types.Container, []byte, string) error
	ExecContainer(*types.Container, *types.Exec, io.Writer) error
	GetContainerStatus(*types.Container) (map[string]string, error)
	IsContainerRunning(*types.Container) (bool, error)
	GetLogs(*types.Container, bool, io.Writer) error
}

// instance is the internal representation of the Backend object.
type instance struct {
	cli       kubernetes.Interface
	cfg       *rest.Config
	initImage string
	namespace string
}

// Config is the structure to instantiate a Backend object
type Config struct {
	// Client is the kubernetes clientset
	Client kubernetes.Interface
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Namespace is the namespace in which all actions are performed
	Namespace string
	// InitImage is the image that is used as init container to prepare vols
	InitImage string
}

// New will return an ContainerFactory instance.
func New(cfg Config) Backend {
	return &instance{
		cli:       cfg.Client,
		cfg:       cfg.RestConfig,
		initImage: cfg.InitImage,
		namespace: cfg.Namespace,
	}
}
