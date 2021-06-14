package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"

	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/exec"
	"github.com/joyrex2001/kubedock/internal/util/portforward"
	"github.com/joyrex2001/kubedock/internal/util/tar"
)

// DeployState describes the state of a deployment.
type DeployState int

const (
	// DeployFailed represents a failed deployment
	DeployFailed DeployState = iota
	// DeployRunning represents a running deployment
	DeployRunning
	// DeployCompleted represents a completed deployment
	DeployCompleted
)

// StartContainer will start given container object in kubernetes and
// waits until it's started, or failed with an error.
func (in *instance) StartContainer(tainr *types.Container) (DeployState, error) {
	match := in.getDeploymentMatchLabels(tainr)

	reqlimits, err := tainr.GetResourceRequirements()
	if err != nil {
		return DeployFailed, err
	}

	// base deploment
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   in.namespace,
			Name:        tainr.ShortID,
			Labels:      in.getLabels(tainr),
			Annotations: in.getAnnotations(tainr),
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
						Image:     tainr.Image,
						Name:      "main",
						Args:      tainr.Cmd,
						Env:       tainr.GetEnvVar(),
						Ports:     in.getContainerPorts(tainr),
						Resources: reqlimits,
					}},
				},
			},
		},
	}
	if tainr.HasVolumes() {
		in.addVolumes(tainr, dep)
	}

	if _, err := in.cli.AppsV1().Deployments(in.namespace).Create(context.TODO(), dep, metav1.CreateOptions{}); err != nil {
		return DeployFailed, err
	}

	if tainr.HasVolumes() {
		if err := in.copyVolumeFolders(tainr); err != nil {
			return DeployFailed, err
		}
	}

	state, err := in.waitReadyState(tainr, in.timeOut)
	if err != nil {
		return state, err
	}

	for _, pp := range tainr.GetContainerTCPPorts() {
		tainr.MapPort(portforward.RandomPort(), pp)
	}

	go func() {
		if err := in.portForward(tainr, tainr.HostPorts); err != nil {
			klog.Errorf("portforward failed: %s", err)
		}
		if err := in.portForward(tainr, tainr.MappedPorts); err != nil {
			klog.Errorf("portforward failed: %s", err)
		}
	}()

	if err := in.CreateServices(tainr); err != nil {
		return state, err
	}

	return state, nil
}

// CreateServices will create k8s service objects for each provided
// external name, mapped with provided hostports ports.
func (in *instance) CreateServices(tainr *types.Container) error {
	for _, svc := range in.getServices(tainr) {
		if _, err := in.cli.CoreV1().Services(in.namespace).Create(context.TODO(), &svc, metav1.CreateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// portForward will create port-forwards for all mapped ports.
func (in *instance) portForward(tainr *types.Container, ports map[int]int) error {
	pods, err := in.getPods(tainr)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("no matching pod found")
	}
	for dst, src := range ports {
		if dst < 0 {
			continue
		}
		stop := make(chan struct{}, 1)
		tainr.AddStopChannel(stop)
		go portforward.ToPod(portforward.Request{
			RestConfig: in.cfg,
			Pod:        pods[0],
			LocalPort:  dst,
			PodPort:    src,
			StopCh:     stop,
			ReadyCh:    make(chan struct{}, 1),
		})
	}
	return nil
}

// getServices will return corev1 services objects for the given
// container definition.
func (in *instance) getServices(tainr *types.Container) []corev1.Service {
	svcs := []corev1.Service{}
	ports := map[int]int{}
	for _, pp := range tainr.GetImageTCPPorts() {
		ports[pp] = pp
	}
	for _, pp := range tainr.GetContainerTCPPorts() {
		ports[pp] = pp
	}
	if tainr.HostPorts != nil {
		for src, dst := range tainr.HostPorts {
			if src <= 0 {
				src = dst
			}
			ports[src] = dst
		}
	}
	if len(ports) == 0 {
		// no ports available, can't create a service without ports
		if len(tainr.NetworkAliases) > 0 {
			klog.Infof("ignoring network aliases %v, no ports mapped", tainr.NetworkAliases)
		}
		return svcs
	}
	valid := regexp.MustCompile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
	for _, alias := range tainr.NetworkAliases {
		if ok := valid.MatchString(alias); !ok {
			klog.Infof("ignoring network alias %s, invalid name", alias)
			continue
		}
		svc := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   in.namespace,
				Name:        alias,
				Labels:      in.getLabels(tainr),
				Annotations: in.getAnnotations(tainr),
			},
			Spec: corev1.ServiceSpec{
				Selector: in.getDeploymentMatchLabels(tainr),
				Ports:    []corev1.ServicePort{},
			},
		}
		for src, dst := range ports {
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
				Name:       fmt.Sprintf("tcp-%d-%d", src, dst),
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(src),
				TargetPort: intstr.IntOrString{IntVal: int32(dst)},
			})
		}
		svcs = append(svcs, svc)
	}
	return svcs
}

// getContainerPorts will return the mapped ports of the container
// as k8s ContainerPorts.
func (in *instance) getContainerPorts(tainr *types.Container) []corev1.ContainerPort {
	res := []corev1.ContainerPort{}
	for _, pp := range tainr.GetContainerTCPPorts() {
		n := fmt.Sprintf("kd-tcp-%d", pp)
		res = append(res, corev1.ContainerPort{ContainerPort: int32(pp), Name: n, Protocol: corev1.ProtocolTCP})
	}
	return res
}

// getLabels will return a map of labels to be added to the container. This
// map contains the labels that link to the container definition, as well
// as additional labels which are used internally by kubedock.
func (in *instance) getLabels(tainr *types.Container) map[string]string {
	l := config.DefaultLabels
	l["kubedock.containerid"] = tainr.ShortID
	return l
}

// getAnnotations will return a map of annotations to be added to the
// container. This map contains the labels as specified in the container
// definition.
func (in *instance) getAnnotations(tainr *types.Container) map[string]string {
	l := tainr.Labels
	if l == nil {
		l = map[string]string{}
	}
	l["kubedock.containername"] = tainr.Name
	return l
}

// getDeploymentMatchLabels will return the map of labels that can be used to
// match running pods for this container.
func (in *instance) getDeploymentMatchLabels(tainr *types.Container) map[string]string {
	return map[string]string{
		"kubedock": tainr.ShortID,
	}
}

// waitReadyState will wait for the deploymemt to be ready.
func (in *instance) waitReadyState(tainr *types.Container, wait int) (DeployState, error) {
	for max := 0; max < wait; max++ {
		dep, err := in.cli.AppsV1().Deployments(in.namespace).Get(context.TODO(), tainr.ShortID, metav1.GetOptions{})
		if err != nil {
			return DeployFailed, err
		}
		if dep.Status.ReadyReplicas > 0 {
			return DeployRunning, nil
		}
		pods, err := in.getPods(tainr)
		if err != nil {
			return DeployFailed, err
		}
		for _, pod := range pods {
			if pod.Status.Phase == corev1.PodFailed {
				return DeployFailed, fmt.Errorf("failed to start container")
			}
			for _, status := range pod.Status.ContainerStatuses {
				term := status.LastTerminationState.Terminated
				if term != nil && term.Reason == "Completed" {
					return DeployCompleted, nil
				}
				if status.RestartCount > 0 {
					return DeployFailed, fmt.Errorf("failed to start container")
				}
			}
		}
		time.Sleep(time.Second)
	}
	return DeployFailed, fmt.Errorf("timeout starting container")
}

// waitInitContainerRunning will wait for a specific container in the
// deployment to be ready.
func (in *instance) waitInitContainerRunning(tainr *types.Container, name string, wait int) error {
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
func (in *instance) addVolumes(tainr *types.Container, dep *appsv1.Deployment) {
	volumes := tainr.GetVolumes()
	dep.Spec.Template.Spec.InitContainers = []corev1.Container{{
		Name:    "setup",
		Image:   in.initImage,
		Command: []string{"sh", "-c", "while [ ! -f /tmp/done ]; do sleep 0.1 ; done"},
	}}

	dep.Spec.Template.Spec.Volumes = []corev1.Volume{}
	mounts := []corev1.VolumeMount{}
	for rm := range volumes {
		id := in.toKubernetesName(rm)
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
func (in *instance) copyVolumeFolders(tainr *types.Container) error {
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
				klog.Errorf("error during tar: %s", err)
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
func (in *instance) signalDone(tainr *types.Container) error {
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
