package types

import (
	"time"
)

// Network describes the details of a network.
type Network struct {
	ID      string
	Name    string
	Created time.Time
}
