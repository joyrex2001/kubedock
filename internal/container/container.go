package container

import (
	"log"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Container describes the details of a container.
type Container struct {
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

// GetShortName will return the a k8s compatible name of the container.
func (co *Container) GetKubernetesName() string {
	n := co.Name
	if n == "" {
		n = co.ID
	}
	if len(n) > 63 {
		return n[:63]
	}
	return n
}

// GetEnvMap will return the environment variables of the container
// as k8s EnvVars.
func (co *Container) GetEnvVar() []corev1.EnvVar {
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

// Delete will delete the Container instance.
func (co *Container) Delete() error {
	return co.db.Delete(co.ID)
}

// Update will update the Container instance.
func (co *Container) Update() error {
	return co.db.Update(co.ID, co)
}
