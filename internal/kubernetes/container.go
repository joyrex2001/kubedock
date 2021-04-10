package kubernetes

import (
	"fmt"

	"github.com/joyrex2001/kubedock/internal/container"
)

func StartContainer(tainr *container.Container) error {
	// return fmt.Errorf("container %s could not be started", tainr.ID)
	return nil
}

func StopContainer(tainr *container.Container) error {
	return fmt.Errorf("container %s could not be stopped", tainr.ID)
}
