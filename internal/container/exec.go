package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Exec is the exec interface to execute arbitrary statements
// within a running container.
type Exec interface {
	Run() error
	Delete() error
	Update() error
}

// ExecObject is the operational implementation of the Exec interace.
type ExecObject struct {
	db keyval.Database
	ID string
}

// Run will execute code inside a running container.
func (eo ExecObject) Run() error {
	// TODO: implement, probably in the kubernetes object.
	return nil
}

// Delete will delete the ExecObject instance.
func (eo ExecObject) Delete() error {
	return eo.db.Delete(eo.ID)
}

// Update will update the ExecObject instance.
func (eo ExecObject) Update() error {
	return eo.db.Update(eo.ID, eo)
}
