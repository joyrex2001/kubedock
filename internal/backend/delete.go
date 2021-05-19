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
	if err := in.deleteServices(tainr.ShortID); err != nil {
		klog.Errorf("error deleting services: %s", err)
	}
	return in.cli.AppsV1().Deployments(in.namespace).Delete(context.TODO(), tainr.ShortID, metav1.DeleteOptions{})
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
			if err := in.deleteServices(dep.Name); err != nil {
				klog.Errorf("error deleting services: %s", err)
			}
			if err := in.cli.AppsV1().Deployments(dep.Namespace).Delete(context.TODO(), dep.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteServicesOlderThan will delete services than are orchestrated
// by kubedock and are older than the given keepmax duration.
func (in *instance) DeleteServicesOlderThan(keepmax time.Duration) error {
	svcs, err := in.cli.CoreV1().Services(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		if svc.ObjectMeta.DeletionTimestamp != nil {
			klog.V(3).Infof("skipping service %v, already in deleting state", svc)
			continue
		}
		old := metav1.NewTime(time.Now().Add(-keepmax))
		if svc.ObjectMeta.CreationTimestamp.Before(&old) {
			klog.V(3).Infof("deleting service: %s", svc.Name)
			if err := in.cli.CoreV1().Services(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// deleteServices will delete k8s service resources which have the
// label kubedock with the given id as value.
func (in *instance) deleteServices(id string) error {
	svcs, err := in.cli.CoreV1().Services(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock.containerid=" + id,
	})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		if err := in.cli.CoreV1().Services(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}
