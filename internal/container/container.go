package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Container is interface that handles the management of container
// objects.
type Container interface {
	GetID() string
	GetName() string
	SetName(string)
	GetImage() string
	SetImage(string)
	GetCmd() []string
	SetCmd([]string)
	GetEnv() []string
	SetEnv([]string)
	GetExposedPorts() map[string]interface{}
	SetExposedPorts(map[string]interface{})
	GetLabels() map[string]string
	SetLabels(map[string]string)
	CreateExec() Exec
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

// CreateExec will create an Exec for the current container.
func (co *Object) CreateExec() Exec {
	// TODO: load exec, delete exec? cascade delete
	return &ExecObject{}
}
