package backend

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// DeleteContainer will delete given container object in kubernetes.
func (in *instance) DeleteContainer(tainr *types.Container) error {
	name := in.getContainerName(tainr)
	return in.cli.AppsV1().Deployments(in.namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// DeleteContainersOlderThan will delete containers than are orchestrated
// by kubedock and are older than the given keepmax duration.
func (in *instance) DeleteContainersOlderThan(keepmax time.Duration) error {
	deps, err := in.cli.AppsV1().Deployments(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, dep := range deps.Items {
		if dep.ObjectMeta.DeletionTimestamp != nil {
			klog.V(3).Infof("skipping deployment %v, already in deleting state", dep)
			continue
		}
		old := metav1.NewTime(time.Now().Add(-keepmax))
		if dep.ObjectMeta.CreationTimestamp.Before(&old) {
			klog.V(3).Infof("deleting deployment: %s", dep.Name)
			if err := in.cli.AppsV1().Deployments(dep.Namespace).Delete(context.TODO(), dep.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}
