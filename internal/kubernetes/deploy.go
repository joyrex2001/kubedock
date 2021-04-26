package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joyrex2001/kubedock/internal/container"
)

// StartContainer will start given container object in kubernetes.
func (in *instance) StartContainer(tainr container.Container) error {
	match := in.getDeploymentMatchLabels(tainr)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.namespace,
			Name:      tainr.GetKubernetesName(),
			Labels:    tainr.GetLabels(), // TODO: add generic label, add ttl annotation, template?)
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: match,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: match,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: tainr.GetImage(),
						Name:  tainr.GetKubernetesName(),
						Args:  tainr.GetCmd(),
						Env:   tainr.GetEnvVar(),
						Ports: in.getContainerPorts(tainr),
					}},
				},
			},
		},
	}

	if _, err := in.cli.AppsV1().Deployments(in.namespace).Create(context.TODO(), dep, metav1.CreateOptions{}); err != nil {
		return err
	}

	// TODO: create port-forward https://github.com/kubernetes/client-go/issues/51

	return nil
}

// getContainerPorts will return the mapped ports of the container
// as k8s ContainerPorts.
func (in *instance) getContainerPorts(tainr container.Container) []corev1.ContainerPort {
	res := []corev1.ContainerPort{}
	for _, pp := range tainr.GetContainerTCPPorts() {
		n := fmt.Sprintf("kd-tcp-%d", pp)
		res = append(res, corev1.ContainerPort{ContainerPort: int32(pp), Name: n, Protocol: corev1.ProtocolTCP})
	}
	return res
}

// getDeploymentMatchLabels will return the map of labels that can be used to match
// running pods for this container.
func (in *instance) getDeploymentMatchLabels(tainr container.Container) map[string]string {
	return map[string]string{
		"app":      tainr.GetKubernetesName(),
		"kubedock": tainr.GetID(),
		"tier":     "kubedock",
	}
}

// GetPodNames will return a list of pods that are spun up for this deployment.
func (in *instance) GetPodNames(tainr container.Container) ([]string, error) {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: in.GetPodsLabelSelector(tainr),
	})
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, p := range pods.Items {
		names = append(names, p.Name)
	}
	return names, nil
}

// GetPodsLabelSelector will return a label selector that can be used to
// uniquely idenitify pods that belong to this deployment.
func (in *instance) GetPodsLabelSelector(tainr container.Container) string {
	return "kubedock=" + tainr.GetID()
}
