package kubernetes

import (
	"log"

	"github.com/joyrex2001/kubedock/internal/container"
)

// Kubernetes is the interface to orchestrate and manage kubernetes objects.
type Kubernetes interface {
	StartContainer(container.Container) error
	DeleteContainer(container.Container) error
}

// instance is the internal representation of the Kubernetes object.
type instance struct {
}

// NewFactory will return an ContainerFactory instance.
func New() Kubernetes {
	return &instance{}
}

// StartContainer will start given container object in kubernetes.
func (in *instance) StartContainer(tainr container.Container) error {
	log.Printf("starting container %s", tainr.GetID())
	return nil
}

// DeleteContainer will delete given container object in kubernetes.
func (in *instance) DeleteContainer(tainr container.Container) error {
	log.Printf("deleting container %s", tainr.GetID())
	return nil
}
