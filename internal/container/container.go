package container

import (
	"log"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Container is interface that handles the management of container
// objects.
type Container interface {
	GetID() string
	GetName() string
	GetKubernetesName() string
	SetName(string)
	GetImage() string
	SetImage(string)
	GetCmd() []string
	SetCmd([]string)
	GetEnv() []string
	GetEnvVar() []corev1.EnvVar
	SetEnv([]string)
	GetExposedPorts() map[string]interface{}
	SetExposedPorts(map[string]interface{})
	MapPort(int, int)
	GetMappedPorts() map[int]int
	GetContainerTCPPorts() []int
	GetLabels() map[string]string
	SetLabels(map[string]string)
	Delete() error
	Update() error
}

// Object is the implementation of the Container interface.
type Object struct {
	db           keyval.Database
	ID           string
	Name         string
	Image        string
	Cmd          []string
	Env          []string
	ExposedPorts map[string]interface{}
	Labels       map[string]string
	MappedPorts  map[int]int
}

// GetID will return the current internal ID of the container.
func (co *Object) GetID() string {
	return co.ID
}

// GetName will return the name of the container.
func (co *Object) GetName() string {
	return co.Name
}

// GetShortName will return the a k8s compatible name of the container.
func (co *Object) GetKubernetesName() string {
	n := co.Name
	if n == "" {
		n = co.ID
	}
	if len(n) > 63 {
		return n[0:63]
	}
	return n
}

// SetName will update the name of the container.
func (co *Object) SetName(name string) {
	co.Name = name
}

// GetImage will return the imagename of the container.
func (co *Object) GetImage() string {
	return co.Image
}

// SetImage will update the imagename of the container.
func (co *Object) SetImage(image string) {
	co.Image = image
}

// GetCmd will return the cmd args of the container.
func (co *Object) GetCmd() []string {
	return co.Cmd
}

// SetCmd will update the cmd args of the container.
func (co *Object) SetCmd(cmd []string) {
	co.Cmd = cmd
}

// GetEnv will return the environment variables of the container.
func (co *Object) GetEnv() []string {
	return co.Env
}

// GetEnvMap will return the environment variables of the container
// as k8s EnvVars.
func (co *Object) GetEnvVar() []corev1.EnvVar {
	env := []corev1.EnvVar{}
	for _, e := range co.Env {
		f := strings.Split(e, "=")
		if len(f) != 2 {
			log.Printf("could not parse env %s", e)
			continue
		}
		env = append(env, corev1.EnvVar{Name: f[0], Value: f[1]})
	}
	return env
}

// SetEnv will update the environment variables of the container.
func (co *Object) SetEnv(env []string) {
	co.Env = env
}

// GetExposedPorts will return the mapped ports of the container.
func (co *Object) GetExposedPorts() map[string]interface{} {
	return co.ExposedPorts
}

// SetExposedPorts will update the mapped ports of the container.
func (co *Object) SetExposedPorts(ports map[string]interface{}) {
	co.ExposedPorts = ports
}

// MapPort will map a pod port to a local port.
func (co *Object) MapPort(pod, local int) {
	if co.MappedPorts == nil {
		co.MappedPorts = map[int]int{}
	}
	co.MappedPorts[pod] = local
}

// GetMappedPorts will return the port mapping setup for the container.
func (co *Object) GetMappedPorts() map[int]int {
	return co.MappedPorts
}

// GetContainerTCPPorts will return a list of all ports that are
// exposed by this container.
func (co *Object) GetContainerTCPPorts() []int {
	ports := []int{}
	for p := range co.ExposedPorts {
		f := strings.Split(p, "/")
		if len(f) != 2 {
			log.Printf("could not parse exposed port %s", p)
			continue
		}
		pp, err := strconv.Atoi(f[0])
		if err != nil {
			log.Printf("could not parse exposed port %s: %s", p, err)
			continue
		}
		if f[1] != "tcp" {
			log.Printf("unsupported protocol %s for port: %d - only tcp is supported", f[1], pp)
			continue
		}
		ports = append(ports, pp)
	}
	return ports
}

// GetLabels will return the labels of the container.
func (co *Object) GetLabels() map[string]string {
	return co.Labels
}

// SetLabels will update the labels of the container.
func (co *Object) SetLabels(labels map[string]string) {
	co.Labels = labels
}

// Delete will delete the ContainerObject instance.
func (co *Object) Delete() error {
	return co.db.Delete(co.ID)
}

// Update will update the ContainerObject instance.
func (co *Object) Update() error {
	return co.db.Update(co.ID, co)
}
