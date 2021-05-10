package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// Container describes the details of a container.
type Container struct {
	ID           string
	Name         string
	Image        string
	Cmd          []string
	Env          []string
	Binds        []string
	ExposedPorts map[string]interface{}
	Networks     map[string]interface{}
	Labels       map[string]string
	MappedPorts  map[int]int
	StopChannels []chan struct{}
	Created      time.Time
}

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

// MapPort will map a pod port to a local port.
func (co *Container) MapPort(pod, local int) {
	if co.MappedPorts == nil {
		co.MappedPorts = map[int]int{}
	}
	co.MappedPorts[pod] = local
}

// GetContainerTCPPorts will return a list of all ports that are
// exposed by this container.
func (co *Container) GetContainerTCPPorts() []int {
	ports := []int{}
	for p := range co.ExposedPorts {
		f := strings.Split(p, "/")
		if len(f) != 2 {
			klog.Errorf("could not parse exposed port %s", p)
			continue
		}
		pp, err := strconv.Atoi(f[0])
		if err != nil {
			klog.Errorf("could not parse exposed port %s: %s", p, err)
			continue
		}
		if f[1] != "tcp" {
			klog.Errorf("unsupported protocol %s for port: %d - only tcp is supported", f[1], pp)
			continue
		}
		ports = append(ports, pp)
	}
	return ports
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
	}
}

// ConnectNetwork will attach a network to the container,
func (co *Container) ConnectNetwork(id string) {
	if co.Networks == nil {
		co.Networks = map[string]interface{}{}
	}
	co.Networks[id] = nil
}

// DisconnectNetwork will detach a network from the container,
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
