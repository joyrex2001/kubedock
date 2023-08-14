package types

import (
	"time"
)

// Exec describes the details of an execute command.
type Exec struct {
	ID          string
	ContainerID string
	Cmd         []string
	TTY         bool
	Stdin       bool
	Stdout      bool
	Stderr      bool
	ExitCode    int
	Created     time.Time
}
