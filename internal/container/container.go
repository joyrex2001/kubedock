package container

import (
	"fmt"

	"github.com/joyrex2001/kubedock/internal/util/uuid"
)

type Container struct {
	ID string
}

func New() *Container {
	id, _ := uuid.New()
	tainr := &Container{
		ID: id,
	}
	return tainr
}

func Load(id string) (*Container, error) {
	return nil, fmt.Errorf("container %s does not exist", id)
}
