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
	StartContainer(*types.Container) (DeployState, error)
	GetContainerStatus(*types.Container) (DeployState, error)
	CreatePortForwards(*types.Container)
	CreateReverseProxies(*types.Container)
	GetPodIP(*types.Container) (string, error)
	DeleteAll() error
	DeleteWithKubedockID(string) error
	DeleteContainer(*types.Container) error
	DeleteOlderThan(time.Duration) error
	CopyFromContainer(tainr *types.Container, path string) ([]byte, error)
	CopyToContainer(*types.Container, []byte, string) error
	ExecContainer(*types.Container, *types.Exec, io.Writer) (int, error)
	GetLogs(*types.Container, bool, int, chan struct{}, io.Writer) error
	GetImageExposedPorts(string) (map[string]struct{}, error)
}

// instance is the internal representation of the Backend object.
type instance struct {
	cli              kubernetes.Interface
	cfg              *rest.Config
	initImage        string
	imagePullSecrets []string
	namespace        string
	timeOut          int
}

// Config is the structure to instantiate a Backend object
type Config struct {
	// Client is the kubernetes clientset
	Client kubernetes.Interface
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Namespace is the namespace in which all actions are performed
	Namespace string
	// ImagePullSecrets is an optional list of image pull secrets that need
	// to be added to the used pod templates
	ImagePullSecrets []string
	// InitImage is the image that is used as init container to prepare vols
	InitImage string
	// TimeOut is the max amount of time to wait until a container started
	TimeOut time.Duration
}

// New will return an Backend instance.
func New(cfg Config) Backend {
	return &instance{
		cli:              cfg.Client,
		cfg:              cfg.RestConfig,
		initImage:        cfg.InitImage,
		namespace:        cfg.Namespace,
		imagePullSecrets: cfg.ImagePullSecrets,
		timeOut:          int(cfg.TimeOut.Seconds()),
	}
}
