package kubernetes

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/joyrex2001/kubedock/internal/container"
	"github.com/joyrex2001/kubedock/internal/util/exec"
	"github.com/joyrex2001/kubedock/internal/util/portforward"
	"github.com/joyrex2001/kubedock/internal/util/tar"
)

// StartContainer will start given container object in kubernetes and
// waits until it's started, or failed with an error.
func (in *instance) StartContainer(tainr *container.Container) error {
	match := in.getDeploymentMatchLabels(tainr)

	// base deploment
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.namespace,
			Name:      tainr.GetKubernetesName(),
			Labels:    tainr.Labels, // TODO: add generic label, add ttl annotation, template?)
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
						Image: tainr.Image,
						Name:  "main",
						Args:  tainr.Cmd,
						Env:   tainr.GetEnvVar(),
						Ports: in.getContainerPorts(tainr),
					}},
				},
			},
		},
	}
	if tainr.HasVolumes() {
		in.addVolumes(tainr, dep)
	}

	if _, err := in.cli.AppsV1().Deployments(in.namespace).Create(context.TODO(), dep, metav1.CreateOptions{}); err != nil {
		return err
	}

	if tainr.HasVolumes() {
		if err := in.copyVolumeFolders(tainr); err != nil {
			return err
		}
	}

	if err := in.waitReadyState(tainr, 30); err != nil {
		return err
	}

	for _, pp := range tainr.GetContainerTCPPorts() {
		tainr.MapPort(pp, portforward.RandomPort())
	}

	go func() {
		err := in.PortForward(tainr)
		if err != nil {
			log.Printf("portforward failed: %s", err)
			return
		}
	}()

	return nil
}

// PortForward will create port-forwards for all mapped ports.
func (in *instance) PortForward(tainr *container.Container) error {
	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no matching pod found")
	}
	for src, dst := range tainr.MappedPorts {
		stream := genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		}
		stop := make(chan struct{}, 1)
		tainr.AddStopChannel(stop)
		go portforward.ToPod(portforward.Request{
			RestConfig: in.cfg,
			Pod:        pods[0],
			LocalPort:  dst,
			PodPort:    src,
			Streams:    stream,
			StopCh:     stop,
			ReadyCh:    make(chan struct{}, 1),
		})
	}
	return nil
}

// getContainerPorts will return the mapped ports of the container
// as k8s ContainerPorts.
func (in *instance) getContainerPorts(tainr *container.Container) []corev1.ContainerPort {
	res := []corev1.ContainerPort{}
	for _, pp := range tainr.GetContainerTCPPorts() {
		n := fmt.Sprintf("kd-tcp-%d", pp)
		res = append(res, corev1.ContainerPort{ContainerPort: int32(pp), Name: n, Protocol: corev1.ProtocolTCP})
	}
	return res
}

// getDeploymentMatchLabels will return the map of labels that can be used to
// match running pods for this container.
func (in *instance) getDeploymentMatchLabels(tainr *container.Container) map[string]string {
	return map[string]string{
		"app":      tainr.GetKubernetesName(),
		"kubedock": tainr.ID,
		"tier":     "kubedock",
	}
}

// getPodsLabelSelector will return a label selector that can be used to
// uniquely idenitify pods that belong to this deployment.
func (in *instance) getPodsLabelSelector(tainr *container.Container) string {
	return "kubedock=" + tainr.ID
}

// waitReadyState will wait for the deploymemt to be ready.
func (in *instance) waitReadyState(tainr *container.Container, wait int) error {
	name := tainr.GetKubernetesName()
	for max := 0; max < wait; max++ {
		dep, err := in.cli.AppsV1().Deployments(in.namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if dep.Status.ReadyReplicas > 0 {
			return nil
		}
		pods, err := in.getPods(tainr)
		if err != nil {
			return err
		}
		for _, pod := range pods {
			if pod.Status.Phase == corev1.PodFailed {
				return fmt.Errorf("failed to start container")
			}
			for _, status := range pod.Status.ContainerStatuses {
				if status.RestartCount > 0 {
					return fmt.Errorf("failed to start container")
				}
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout starting container")
}

// waitReadyContainer will wait for a specific container in the deployment
// to be ready.
func (in *instance) waitInitContainerRunning(tainr *container.Container, name string, wait int) error {
	for max := 0; max < wait; max++ {
		pods, err := in.getPods(tainr)
		if err != nil {
			return err
		}
		for _, pod := range pods {
			if pod.Status.Phase == corev1.PodFailed {
				return fmt.Errorf("failed to start container")
			}
			for _, status := range pod.Status.InitContainerStatuses {
				if status.Name != name {
					continue
				}
				if status.State.Running != nil {
					return nil
				}
			}
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout starting container")
}

// addVolumes will add an init-container "setup" and creates volumes and
// volume mounts in both the init container and "main" container in order
// to copy data before the container is started.
func (in *instance) addVolumes(tainr *container.Container, dep *appsv1.Deployment) {
	volumes := tainr.GetVolumes()
	dep.Spec.Template.Spec.InitContainers = []corev1.Container{{
		Name:    "setup",
		Image:   in.initImage,
		Command: []string{"sh", "-c", "while [ ! -f /tmp/done ]; do sleep 0.1 ; done"},
	}}

	dep.Spec.Template.Spec.Volumes = []corev1.Volume{}
	mounts := []corev1.VolumeMount{}
	for rm := range volumes {
		id := in.getVolumeId(rm)
		dep.Spec.Template.Spec.Volumes = append(dep.Spec.Template.Spec.Volumes,
			corev1.Volume{Name: id, VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}})
		mounts = append(mounts, corev1.VolumeMount{Name: id, MountPath: rm})
	}
	dep.Spec.Template.Spec.Containers[0].VolumeMounts = mounts
	dep.Spec.Template.Spec.InitContainers[0].VolumeMounts = mounts
}

// copyVolumeFolders will copy the configured volumes of the container to
// the running init container, and signal the init container when finished
// with copying.
func (in *instance) copyVolumeFolders(tainr *container.Container) error {
	if err := in.waitInitContainerRunning(tainr, "setup", 30); err != nil {
		return err
	}

	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}

	volumes := tainr.GetVolumes()
	for rm, src := range volumes {
		reader, writer := io.Pipe()
		go func() {
			defer writer.Close()
			if err := tar.PackFolder(src, writer); err != nil {
				log.Printf("error during tar: %s", err)
				return
			}
		}()
		if err := exec.RemoteCmd(exec.Request{
			Client:     in.cli,
			RestConfig: in.cfg,
			Pod:        pods[0],
			Container:  "setup",
			Cmd:        []string{"tar", "-xf", "-", "-C", rm},
			Stdin:      reader,
		}); err != nil {
			return err
		}
	}

	return in.signalDone(tainr)
}

// signalDone will signal the prepare init container to exit.
func (in *instance) signalDone(tainr *container.Container) error {
	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}
	return exec.RemoteCmd(exec.Request{
		Client:     in.cli,
		RestConfig: in.cfg,
		Pod:        pods[0],
		Container:  "setup",
		Cmd:        []string{"touch", "/tmp/done"},
		Stderr:     os.Stderr,
	})
}

// getVolumeId creates an id to use for the volume mapping
func (in *instance) getVolumeId(path string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9-]`)
	id := re.ReplaceAllString(path, ``)
	if len(id) > 63 {
		id = id[len(id)-63:]
	}
	return id
}

// getPods will return a list of pods that are spun up for this deployment.
func (in *instance) getPods(tainr *container.Container) ([]corev1.Pod, error) {
	pods, err := in.cli.CoreV1().Pods(in.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: in.getPodsLabelSelector(tainr),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}
