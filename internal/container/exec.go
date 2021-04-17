package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

type Exec interface {
	Run() error
	Delete() error
	Update() error
}

type ExecObject struct {
	db keyval.Database
	ID string
}

func (eo ExecObject) Run() error {
	return nil
}

func (eo ExecObject) Delete() error {
	return eo.db.Delete(eo.ID)
}

func (eo ExecObject) Update() error {
	return eo.db.Update(eo.ID, eo)
}
