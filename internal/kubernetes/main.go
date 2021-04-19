package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joyrex2001/kubedock/internal/container"
)

// Kubernetes is the interface to orchestrate and manage kubernetes objects.
type Kubernetes interface {
	StartContainer(container.Container) error
	GetContainerStatus(container.Container) (map[string]string, error)
	DeleteContainer(container.Container) error
	ExecContainer(container.Container, container.Exec) error
	GetExecStatus(container.Exec) (map[string]string, error)
}

// instance is the internal representation of the Kubernetes object.
type instance struct {
	cli       *kubernetes.Clientset
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

// GetExecStatus will return current status of given exec object in kubernetes.
func (in *instance) GetExecStatus(exec container.Exec) (map[string]string, error) {
	return nil, nil
}
