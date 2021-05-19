package types

import (
	"time"
)

// Network describes the details of a network.
type Network struct {
	ID      string
	ShortID string
	Name    string
	Created time.Time
}
