package kubernetes

import (
	"github.com/joyrex2001/kubedock/internal/container"
)

func StartContainer(tainr *container.Container) error {
	// return fmt.Errorf("container %s could not be started", tainr.ID)
	return nil
}

func DeleteContainer(tainr *container.Container) error {
	// return fmt.Errorf("container %s could not be deleted", tainr.ID)
	return nil
}
