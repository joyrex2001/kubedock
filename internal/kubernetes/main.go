package kubernetes

import (
	"log"

	"k8s.io/client-go/kubernetes"

	"github.com/joyrex2001/kubedock/internal/container"
)

// Kubernetes is the interface to orchestrate and manage kubernetes objects.
type Kubernetes interface {
	StartContainer(container.Container) error
	GetContainerStatus(container.Container) (map[string]string, error)
	DeleteContainer(container.Container) error
	ExecContainer(container.Exec) error
	GetExecStatus(container.Exec) (map[string]string, error)
}

// instance is the internal representation of the Kubernetes object.
type instance struct {
	cli       *kubernetes.Clientset
	namespace string
}

// NewFactory will return an ContainerFactory instance.
func New(cli *kubernetes.Clientset, namespace string) Kubernetes {
	return &instance{
		cli:       cli,
		namespace: namespace,
	}
}

// GetContainerStatus will return current status of given exec object in kubernetes.
func (in *instance) GetContainerStatus(tainr container.Container) (map[string]string, error) {
	return nil, nil
}

// DeleteContainer will delete given container object in kubernetes.
func (in *instance) DeleteContainer(tainr container.Container) error {
	log.Printf("deleting container %s", tainr.GetID())
	return nil
}

// ExecContainer will execute given exec object in kubernetes.
func (in *instance) ExecContainer(exec container.Exec) error {
	log.Printf("exec %s", exec.GetID())
	return nil
}

// GetExecStatus will return current status of given exec object in kubernetes.
func (in *instance) GetExecStatus(exec container.Exec) (map[string]string, error) {
	return nil, nil
}
