package kubernetes

import (
	"io"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// Kubernetes is the interface to orchestrate and manage kubernetes objects.
type Kubernetes interface {
	StartContainer(*types.Container) error
	DeleteContainer(*types.Container) error
	CopyToContainer(*types.Container, []byte, string) error
	ExecContainer(*types.Container, *types.Exec, io.Writer) error
	GetContainerStatus(*types.Container) (map[string]string, error)
	IsContainerRunning(*types.Container) (bool, error)
	GetLogs(*types.Container, bool, io.Writer) error
}

// instance is the internal representation of the Kubernetes object.
type instance struct {
	cli       kubernetes.Interface
	cfg       *rest.Config
	initImage string
	namespace string
}

// New will return an ContainerFactory instance.
func New(cfg *rest.Config, cli *kubernetes.Clientset, namespace string) Kubernetes {
	return &instance{
		cli:       cli,
		cfg:       cfg,
		initImage: "busybox:latest", // TODO: configureable, default to kubedock image
		namespace: namespace,
	}
}
