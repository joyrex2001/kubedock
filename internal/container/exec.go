package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Exec describes the details of an execute command.
type Exec struct {
	db          keyval.Database
	ID          string
	ContainerID string
	Cmd         []string
	Stdout      bool
	Stderr      bool
}

// Delete will delete the ExecObject instance.
func (eo *Exec) Delete() error {
	return eo.db.Delete(eo.ID)
}

// Update will update the ExecObject instance.
func (eo *Exec) Update() error {
	return eo.db.Update(eo.ID, eo)
}
