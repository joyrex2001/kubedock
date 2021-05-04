package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// GetContainerStatus will return current status of given exec object in kubernetes.
func (in *instance) GetContainerStatus(tainr *types.Container) (map[string]string, error) {
	name := tainr.GetKubernetesName()
	dep, err := in.cli.AppsV1().Deployments(in.namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if dep.Status.ReadyReplicas > 0 {
		return map[string]string{
			"Status":  "healthy",
			"Running": "running",
		}, nil
	}
	return map[string]string{
		"Status":  "unhealthy",
		"Running": "created",
	}, nil
}

// IsContainerRunning will return true if the container is in running state.
func (in *instance) IsContainerRunning(tainr *types.Container) (bool, error) {
	status, err := in.GetContainerStatus(tainr)
	if err != nil {
		return false, err
	}
	return status["Running"] == "running", nil
}
