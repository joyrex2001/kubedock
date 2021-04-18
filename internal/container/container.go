package container

import (
	"fmt"
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
	GetContainerPorts() []corev1.ContainerPort
	SetExposedPorts(map[string]interface{})
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

// GetContainerPorts will return the mapped ports of the container
// as k8s ContainerPorts.
func (co *Object) GetContainerPorts() []corev1.ContainerPort {
	ports := []corev1.ContainerPort{}
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
		n := fmt.Sprintf("kd-%s-%d", f[1], pp)
		ports = append(ports, corev1.ContainerPort{ContainerPort: int32(pp), Name: n})
	}
	return ports
}

// SetExposedPorts will update the mapped ports of the container.
func (co *Object) SetExposedPorts(ports map[string]interface{}) {
	co.ExposedPorts = ports
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
