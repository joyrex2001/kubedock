package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Factory is the factory interface to create and load container
// and related objects.
type Factory interface {
	Create() (*Container, error)
	Load(string) (*Container, error)
	CreateExec(string) (*Exec, error)
	LoadExec(string) (*Exec, error)
}

// instance is the internal representation of the Factory object.
type instance struct {
	db keyval.Database
}

// NewFactory will return an ContainerFactory instance.
func NewFactory(kv keyval.Database) Factory {
	return &instance{kv}
}

// Create will create fresh Container objects and will return an
// error if failed.
func (f instance) Create() (*Container, error) {
	res := &Container{
		db: f.db,
	}
	id, err := f.db.Create(res)
	if err != nil {
		return nil, err
	}
	res.ID = id
	return res, nil
}

// Load will return an existing Container object specified with
// the given id. If the Container object does not exist it will
// return an error.
func (f instance) Load(id string) (*Container, error) {
	x, err := f.db.Read(id)
	if err != nil {
		return nil, err
	}
	res := x.(*Container)
	res.ID = id
	return res, nil
}

// CreateExec will create fresh Exec objects for given container
// and will return an error if failed.
func (f instance) CreateExec(containerId string) (*Exec, error) {
	res := &Exec{
		db:          f.db,
		ContainerID: containerId,
	}
	id, err := f.db.Create(res)
	if err != nil {
		return nil, err
	}
	res.ID = id
	return res, nil
}

// Load will return an existing Exec object specified with
// the given id. If the Exec object does not exist it will
// return an error.
func (f instance) LoadExec(id string) (*Exec, error) {
	x, err := f.db.Read(id)
	if err != nil {
		return nil, err
	}
	res := x.(*Exec)
	res.ID = id
	return res, nil
}
