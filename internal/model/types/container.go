package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

// Container describes the details of a container.
type Container struct {
	ID             string
	ShortID        string
	Name           string
	Image          string
	Labels         map[string]string
	Cmd            []string
	Env            []string
	Binds          []string
	HostIP         string
	ExposedPorts   map[string]interface{}
	ImagePorts     map[string]interface{}
	HostPorts      map[int]int
	MappedPorts    map[int]int
	Networks       map[string]interface{}
	NetworkAliases []string
	StopChannels   []chan struct{}
	AttachChannels []chan struct{}
	Running        bool
	Completed      bool
	Failed         bool
	Stopped        bool
	Killed         bool
	Created        time.Time
}

const (
	// LabelRequestCPU is the label to be use to specify cpu request/limits
	LabelRequestCPU = "com.joyrex2001.kubedock.request-cpu"
	// LabelRequestMemory is the label to be use to specify memory request/limits
	LabelRequestMemory = "com.joyrex2001.kubedock.request-memory"
)

// GetEnvVar will return the environment variables of the container
// as k8s EnvVars.
func (co *Container) GetEnvVar() []corev1.EnvVar {
	env := []corev1.EnvVar{}
	for _, e := range co.Env {
		f := strings.Split(e, "=")
		if len(f) != 2 {
			klog.Errorf("could not parse env %s", e)
			continue
		}
		env = append(env, corev1.EnvVar{Name: f[0], Value: f[1]})
	}
	return env
}

// GetResourceRequirements will return a k8s request/limits configuration
// based on the LabelRequestCPU and LabelRequestMemory labels set on the
// container.
func (co *Container) GetResourceRequirements() (corev1.ResourceRequirements, error) {
	req := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	for typ, labl := range map[string]string{"cpu": LabelRequestCPU, "memory": LabelRequestMemory} {
		rls, ok := co.Labels[labl]
		if !ok {
			continue
		}

		var r, l string
		rl := strings.Split(strings.ReplaceAll(rls, " ", ""), ",")
		if len(rl) == 0 || len(rl) > 2 {
			return req, fmt.Errorf("invalid resource requirement: %s", rls)
		}
		r = rl[0]
		if len(rl) == 2 {
			l = rl[1]
		}
		if r == "" && l != "" {
			r = l
		}

		rq, err := resource.ParseQuantity(r)
		if err != nil {
			return req, err
		}
		req.Requests[corev1.ResourceName(typ)] = rq

		if l != "" {
			lt, err := resource.ParseQuantity(l)
			if err != nil {
				return req, err
			}
			req.Limits[corev1.ResourceName(typ)] = lt
		}
	}
	return req, nil
}

// MapPort will map a pod port to a local port.
func (co *Container) MapPort(pod, local int) {
	if co.MappedPorts == nil {
		co.MappedPorts = map[int]int{}
	}
	co.MappedPorts[pod] = local
}

// AddHostPort will add a predefined port mapping.
func (co *Container) AddHostPort(src string, dst string) error {
	var err error
	var sp, dp int

	dp, err = co.getTCPPort(dst)
	if err != nil {
		return err
	}

	if src != "" {
		sp, err = strconv.Atoi(src)
		if err != nil {
			return fmt.Errorf("could not parse exposed port %s: %s", dst, err)
		}
	} else {
		sp = -dp
	}

	if co.HostPorts == nil {
		co.HostPorts = map[int]int{}
	}
	co.HostPorts[sp] = dp

	return nil
}

// GetContainerTCPPorts will return a list of all ports that are
// exposed by this container.
func (co *Container) GetContainerTCPPorts() []int {
	return co.getTCPPorts(co.ExposedPorts)
}

// GetImageTCPPorts will return a list of all ports that are
// exposed by the image.
func (co *Container) GetImageTCPPorts() []int {
	return co.getTCPPorts(co.ImagePorts)
}

// GetServicePorts will return a list of ports and their mapping as they
// should be applied on a k8s service.
func (co *Container) GetServicePorts() map[int]int {
	ports := map[int]int{}
	for _, pp := range co.GetImageTCPPorts() {
		ports[pp] = pp
	}
	for _, pp := range co.GetContainerTCPPorts() {
		ports[pp] = pp
	}
	if co.HostPorts != nil {
		for src, dst := range co.HostPorts {
			if src <= 0 {
				src = dst
			}
			ports[src] = dst
		}
	}
	return ports
}

// getTCPPorts will return a list of all tcp ports in given map.
func (co *Container) getTCPPorts(ports map[string]interface{}) []int {
	res := []int{}
	if ports == nil {
		return res
	}
	for p := range ports {
		pp, err := co.getTCPPort(p)
		if err != nil {
			klog.Errorf("could not parse exposed port %s", p)
			continue
		}
		res = append(res, pp)
	}
	return res
}

// getTCPPort will convert a "9000/tcp" string to the port.
func (co *Container) getTCPPort(p string) (int, error) {
	f := strings.Split(p, "/")
	if len(f) != 2 {
		return 0, fmt.Errorf("could not parse exposed port %s", p)
	}
	pp, err := strconv.Atoi(f[0])
	if err != nil {
		return 0, fmt.Errorf("could not parse exposed port %s: %s", p, err)
	}
	if f[1] != "tcp" {
		return 0, fmt.Errorf("unsupported protocol %s for port: %d - only tcp is supported", f[1], pp)
	}
	return pp, nil
}

// GetVolumes will return a map of volumes that should be mounted on the
// target container. The key is the target location, and the value is the
// local location.
func (co *Container) GetVolumes() map[string]string {
	mounts := map[string]string{}
	for _, bind := range co.Binds {
		f := strings.Split(bind, ":")
		mounts[f[1]] = f[0]
	}
	return mounts
}

// HasVolumes will return true if the container has volumes configured.
func (co *Container) HasVolumes() bool {
	return len(co.Binds) > 0
}

// AddStopChannel will add channels that should be notified when
// SignalStop is called.
func (co *Container) AddStopChannel(stop chan struct{}) {
	if co.StopChannels == nil {
		co.StopChannels = []chan struct{}{}
	}
	co.StopChannels = append(co.StopChannels, stop)
}

// SignalStop will signal all stop channels.
func (co *Container) SignalStop() {
	for _, stop := range co.StopChannels {
		stop <- struct{}{}
		close(stop)
	}
	co.StopChannels = []chan struct{}{}
}

// AddAttachChannel will add channels that should be notified when
// SignalDetach is called.
func (co *Container) AddAttachChannel(stop chan struct{}) {
	if co.AttachChannels == nil {
		co.AttachChannels = []chan struct{}{}
	}
	co.AttachChannels = append(co.AttachChannels, stop)
}

// SignalDetach will signal all stop channels.
func (co *Container) SignalDetach() {
	for _, stop := range co.AttachChannels {
		stop <- struct{}{}
		close(stop)
	}
	co.AttachChannels = []chan struct{}{}
}

// ConnectNetwork will attach a network to the container.
func (co *Container) ConnectNetwork(id string) {
	if co.Networks == nil {
		co.Networks = map[string]interface{}{}
	}
	co.Networks[id] = nil
}

// DisconnectNetwork will detach a network from the container.
func (co *Container) DisconnectNetwork(id string) error {
	if id == "bridge" {
		return fmt.Errorf("can't delete bridge network")
	}
	if _, ok := co.Networks[id]; !ok {
		return fmt.Errorf("container is not connected to network %s", id)
	}
	delete(co.Networks, id)
	return nil
}

// Match will match given type with given key value pair.
func (co *Container) Match(typ string, key string, val string) bool {
	if typ != "label" {
		return true
	}
	v, ok := co.Labels[key]
	if !ok {
		return false
	}
	return v == val
}

// StateString returns a string that describes the state.
func (co *Container) StateString() string {
	if co.Running {
		return "Up"
	}
	if co.Stopped || co.Killed {
		return "Dead"
	}
	if co.Failed {
		return "Dead"
	}
	if co.Completed {
		return "Exited"
	}
	return "Created"
}

// StatusString returns a string that describes the status.
func (co *Container) StatusString() string {
	if co.Running {
		return "healthy"
	}
	return "unhealthy"
}
