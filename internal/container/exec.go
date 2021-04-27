package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Exec is the exec interface to execute arbitrary statements
// within a running container.
type Exec interface {
	GetID() string
	GetContainerID() string
	GetCmd() []string
	SetCmd([]string)
	GetStdout() bool
	SetStdout(flag bool)
	GetStderr() bool
	SetStderr(flag bool)
	Delete() error
	Update() error
}

// ExecObject is the operational implementation of the Exec interace.
type ExecObject struct {
	db          keyval.Database
	ID          string
	ContainerID string
	Cmd         []string
	Stdout      bool
	Stderr      bool
}

// GetID will return the current internal ID of the exec.
func (eo *ExecObject) GetID() string {
	return eo.ID
}

// GetContainerID will return the ID of the container for this exec object.
func (eo *ExecObject) GetContainerID() string {
	return eo.ContainerID
}

// GetCmd will return the cmd args of the exec.
func (eo *ExecObject) GetCmd() []string {
	return eo.Cmd
}

// SetCmd will update the cmd args of the exec.
func (eo *ExecObject) SetCmd(cmd []string) {
	eo.Cmd = cmd
}

// GetStdout will return the stdout flag of the exec.
func (eo *ExecObject) GetStdout() bool {
	return eo.Stdout
}

// SetStdout will update the stdout flag of the exec.
func (eo *ExecObject) SetStdout(flag bool) {
	eo.Stdout = flag
}

// GetStdout will return the stdout flag of the exec.
func (eo *ExecObject) GetStderr() bool {
	return eo.Stderr
}

// SetStdout will update the stdout flag of the exec.
func (eo *ExecObject) SetStderr(flag bool) {
	eo.Stderr = flag
}

// Delete will delete the ExecObject instance.
func (eo ExecObject) Delete() error {
	return eo.db.Delete(eo.ID)
}

// Update will update the ExecObject instance.
func (eo ExecObject) Update() error {
	return eo.db.Update(eo.ID, eo)
}
