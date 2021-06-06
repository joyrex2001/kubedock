package types

import (
	"time"
)

// Exec describes the details of an execute command.
type Exec struct {
	ID          string
	ContainerID string
	Cmd         []string
	Stdout      bool
	Stderr      bool
	ExitCode    int
	Created     time.Time
}
