package kubernetes

import (
	"io"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joyrex2001/kubedock/internal/container"
)

// Kubernetes is the interface to orchestrate and manage kubernetes objects.
type Kubernetes interface {
	StartContainer(*container.Container) error
	DeleteContainer(*container.Container) error
	CopyToContainer(*container.Container, []byte, string) error
	ExecContainer(*container.Container, *container.Exec, io.Writer) error
	GetContainerStatus(*container.Container) (map[string]string, error)
	IsContainerRunning(*container.Container) (bool, error)
	GetLogs(*container.Container, bool, io.Writer) error
}

// instance is the internal representation of the Kubernetes object.
type instance struct {
	cli       kubernetes.Interface
	cfg       *rest.Config
	namespace string
}

// NewFactory will return an ContainerFactory instance.
func New(cfg *rest.Config, cli *kubernetes.Clientset, namespace string) Kubernetes {
	return &instance{
		cli:       cli,
		cfg:       cfg,
		namespace: namespace,
	}
}
