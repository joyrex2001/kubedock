package kubernetes

import (
	"fmt"

	"github.com/joyrex2001/donk/internal/container"
)

func StartContainer(tainr *container.Container) error {
	return fmt.Errorf("container %s could not be started", tainr.ID)
}

func StopContainer(tainr *container.Container) error {
	return fmt.Errorf("container %s could not be stopped", tainr.ID)
}
