package backend

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// Status describes the current status of a running container.
type Status struct {
	Replicas int32
	Created  time.Time
}

// StateString returns a string that describes the state.
func (s *Status) StateString() string {
	if s.Replicas > 0 {
		return "Up"
	}
	return "Created"
}

// StatusString returns a string that describes the status.
func (s *Status) StatusString() string {
	if s.Replicas > 0 {
		return "healthy"
	}
	return "unhealthy"
}

// GetContainerStatus will return current status of given exec object in kubernetes.
func (in *instance) GetContainerStatus(tainr *types.Container) (*Status, error) {
	dep, err := in.cli.AppsV1().Deployments(in.namespace).Get(context.TODO(), tainr.ShortID, metav1.GetOptions{})
	if err != nil {
		return &Status{}, err
	}
	return &Status{
		Replicas: dep.Status.ReadyReplicas,
		Created:  dep.ObjectMeta.CreationTimestamp.Time,
	}, nil
}

// IsContainerRunning will return true if the container is in running state.
func (in *instance) IsContainerRunning(tainr *types.Container) (bool, error) {
	s, err := in.GetContainerStatus(tainr)
	if err != nil {
		return false, err
	}
	return s.Replicas > 0, nil
}
