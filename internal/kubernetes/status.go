package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/container"
)

// GetContainerStatus will return current status of given exec object in kubernetes.
func (in *instance) GetContainerStatus(tainr container.Container) (map[string]string, error) {
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
