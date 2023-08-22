package backend

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

// DeleteAll will delete all resources that kubedock=true
func (in *instance) DeleteAll() error {
	ok := true
	if err := in.deleteServices("kubedock=true"); err != nil {
		klog.Errorf("error deleting services: %s", err)
		ok = false
	}
	if err := in.deleteConfigMaps("kubedock=true"); err != nil {
		klog.Errorf("error deleting configmaps: %s", err)
		ok = false
	}
	if err := in.deletePods("kubedock=true"); err != nil {
		klog.Errorf("error deleting pods: %s", err)
		ok = false
	}
	if !ok {
		return fmt.Errorf("failed deleting all containers")
	}
	return nil
}

// DeleteWithKubedockID will delete all resources that have given kubedock.id
func (in *instance) DeleteWithKubedockID(id string) error {
	ok := true
	if err := in.deleteServices("kubedock.id=" + id); err != nil {
		klog.Errorf("error deleting services: %s", err)
		ok = false
	}
	if err := in.deleteConfigMaps("kubedock.id=" + id); err != nil {
		klog.Errorf("error deleting configmaps: %s", err)
		ok = false
	}
	if err := in.deletePods("kubedock.id=" + id); err != nil {
		klog.Errorf("error deleting pods: %s", err)
		ok = false
	}
	if !ok {
		return fmt.Errorf("failed deleting container %s", id)
	}
	return nil
}

// DeleteContainer will delete given container object in kubernetes.
func (in *instance) DeleteContainer(tainr *types.Container) error {
	ok := true
	if err := in.deleteServices("kubedock.containerid=" + tainr.ShortID); err != nil {
		klog.Errorf("error deleting services: %s", err)
		ok = false
	}
	if err := in.deleteConfigMaps("kubedock.containerid=" + tainr.ShortID); err != nil {
		klog.Errorf("error deleting configmaps: %s", err)
		ok = false
	}
	if err := in.deletePods("kubedock.containerid=" + tainr.ShortID); err != nil {
		klog.Errorf("error deleting pods: %s", err)
		ok = false
	}
	if !ok {
		return fmt.Errorf("failed deleting container %s", tainr.ShortID)
	}
	return nil
}

// DeleteOlderThan will delete all kubedock created resources older
// than the given keepmax duration.
func (in *instance) DeleteOlderThan(keepmax time.Duration) error {
	if err := in.DeleteContainersOlderThan(keepmax); err != nil {
		return err
	}
	if err := in.DeleteConfigMapsOlderThan(keepmax); err != nil {
		return err
	}
	if err := in.DeletePodsOlderThan(keepmax); err != nil {
		return err
	}
	return in.DeleteServicesOlderThan(keepmax)
}

// DeleteContainersOlderThan will delete containers than are orchestrated
// by kubedock and are older than the given keepmax duration.
func (in *instance) DeleteContainersOlderThan(keepmax time.Duration) error {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if in.isOlderThan(pod.ObjectMeta, keepmax) {
			klog.V(3).Infof("deleting pod: %s", pod.Name)
			if err := in.deleteServices("kubedock.containerid=" + pod.Name); err != nil {
				klog.Errorf("error deleting services: %s", err)
			}
			if err := in.deleteConfigMaps("kubedock.containerid=" + pod.Name); err != nil {
				klog.Errorf("error deleting configmaps: %s", err)
			}
			if err := in.cli.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{}); err != nil {
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
		if in.isOlderThan(svc.ObjectMeta, keepmax) {
			klog.V(3).Infof("deleting service: %s", svc.Name)
			if err := in.cli.CoreV1().Services(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteConfigMapsOlderThan will delete configmaps than are orchestrated
// by kubedock and are older than the given keepmax duration.
func (in *instance) DeleteConfigMapsOlderThan(keepmax time.Duration) error {
	svcs, err := in.cli.CoreV1().ConfigMaps(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		if in.isOlderThan(svc.ObjectMeta, keepmax) {
			klog.V(3).Infof("deleting service: %s", svc.Name)
			if err := in.cli.CoreV1().ConfigMaps(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeletePodsOlderThan will delete pods than are orchestrated by kubedock
// and are older than the given keepmax duration.
func (in *instance) DeletePodsOlderThan(keepmax time.Duration) error {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock=true",
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if in.isOlderThan(pod.ObjectMeta, keepmax) {
			klog.V(3).Infof("deleting pod: %s", pod.Name)
			background := metav1.DeletePropagationBackground
			if err := in.cli.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
				PropagationPolicy: &background,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// isOlderThan will check if given resource metadata has an older timestamp
// compared to given keepmax duration
func (in *instance) isOlderThan(met metav1.ObjectMeta, keepmax time.Duration) bool {
	if met.DeletionTimestamp != nil {
		klog.V(3).Infof("ignoring %v, already in deleting state", met)
		return false
	}
	old := metav1.NewTime(time.Now().Add(-keepmax))
	return met.CreationTimestamp.Before(&old)
}

// deleteServices will delete k8s service resources which match the
// given label selector.
func (in *instance) deleteServices(selector string) error {
	svcs, err := in.cli.CoreV1().Services(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
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

// deleteConfigMaps will delete k8s configmap resources which match the
// given label selector.
func (in *instance) deleteConfigMaps(selector string) error {
	svcs, err := in.cli.CoreV1().ConfigMaps(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		if err := in.cli.CoreV1().ConfigMaps(svc.Namespace).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// deletePods will delete k8s pod resources which match the given label
// selector.
func (in *instance) deletePods(selector string) error {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if err := in.cli.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// WatchDeleteContainer will return a channel which will be signalled when
// the given container is actually deleted from kubernetes.
func (in *instance) WatchDeleteContainer(tainr *types.Container, timeout time.Duration) (chan struct{}, error) {
	delch := make(chan struct{}, 1)

	watcher, err := in.cli.CoreV1().Pods(in.namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: "kubedock.containerid=" + tainr.ShortID,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			tmr := time.NewTimer(timeout)
			select {
			case event := <-watcher.ResultChan():
				if event.Type == watch.Deleted {
					close(delch)
					watcher.Stop()
					return
				}
			case <-tmr.C:
				close(delch)
				watcher.Stop()
				return
			}
		}
	}()

	return delch, nil
}
