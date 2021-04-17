package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

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

type ContainerObject struct {
	db           keyval.Database
	ID           string
	Name         string
	Image        string
	Cmd          []string
	Env          []string
	ExposedPorts map[string]interface{}
	Labels       map[string]string
}

func (co *ContainerObject) GetID() string {
	return co.ID
}

func (co *ContainerObject) GetName() string {
	return co.Name
}

func (co *ContainerObject) SetName(name string) {
	co.Name = name
}

func (co *ContainerObject) GetImage() string {
	return co.Image
}

func (co *ContainerObject) SetImage(image string) {
	co.Image = image
}

func (co *ContainerObject) GetCmd() []string {
	return co.Cmd
}

func (co *ContainerObject) SetCmd(cmd []string) {
	co.Cmd = cmd
}

func (co *ContainerObject) GetEnv() []string {
	return co.Env
}

func (co *ContainerObject) SetEnv(env []string) {
	co.Env = env
}

func (co *ContainerObject) GetExposedPorts() map[string]interface{} {
	return co.ExposedPorts
}

func (co *ContainerObject) SetExposedPorts(ports map[string]interface{}) {
	co.ExposedPorts = ports
}

func (co *ContainerObject) GetLabels() map[string]string {
	return co.Labels
}

func (co *ContainerObject) SetLabels(labels map[string]string) {
	co.Labels = labels
}

func (co *ContainerObject) CreateExec() Exec {
	// TODO: load exec, delete exec? cascade delete
	return &ExecObject{}
}

func (co *ContainerObject) Delete() error {
	return co.db.Delete(co.ID)
}

func (co *ContainerObject) Update() error {
	return co.db.Update(co.ID, co)
}
