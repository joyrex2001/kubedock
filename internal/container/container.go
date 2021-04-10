package container

import (
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"

	"github.com/joyrex2001/kubedock/internal/util/uuid"
)

var db *cache.Cache

func init() {
	db = cache.New(cache.DefaultExpiration, 10*time.Minute)
}

type Container struct {
	ID           string
	Name         string
	Image        string
	ExposedPorts map[string]interface{}
	Labels       map[string]string
}

func New(name, image string, ports map[string]interface{}, labels map[string]string) *Container {
	id, _ := uuid.New()
	tainr := &Container{
		ID:           id,
		Name:         name,
		Image:        image,
		ExposedPorts: ports,
		Labels:       labels,
	}
	db.Set(id, tainr, cache.DefaultExpiration)
	return tainr
}

func Load(id string) (*Container, error) {
	x, ok := db.Get(id)
	if !ok {
		return nil, fmt.Errorf("container %s does not exist", id)
	}
	return x.(*Container), nil
}
