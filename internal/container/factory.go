package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

// Factory is the factory interface to create and load container objects.
type Factory interface {
	Create() (Container, error)
	Load(string) (Container, error)
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
func (f instance) Create() (Container, error) {
	res := &Object{
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
func (f instance) Load(id string) (Container, error) {
	x, err := f.db.Read(id)
	if err != nil {
		return nil, err
	}
	res := x.(*Object)
	res.ID = id
	return res, nil
}
