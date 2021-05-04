package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// DeleteContainer will delete given container object in kubernetes.
func (in *instance) DeleteContainer(tainr *types.Container) error {
	name := tainr.GetKubernetesName()
	return in.cli.AppsV1().Deployments(in.namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
