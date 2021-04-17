package container

import (
	"github.com/joyrex2001/kubedock/internal/util/keyval"
)

type ContainerFactory interface {
	Create() (Container, error)
	Load(string) (Container, error)
}

type instance struct {
	db keyval.Database
}

func NewFactory(kv keyval.Database) ContainerFactory {
	return &instance{kv}
}

func (f instance) Create() (Container, error) {
	res := &ContainerObject{
		db: f.db,
	}
	id, err := f.db.Create(res)
	if err != nil {
		return nil, err
	}
	res.ID = id
	return res, nil
}

func (f instance) Load(id string) (Container, error) {
	x, err := f.db.Read(id)
	if err != nil {
		return nil, err
	}
	res := x.(*ContainerObject)
	res.ID = id
	return res, nil
}
