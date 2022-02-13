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

// IsPredefined will return if the network is a pre-defined system network.
func (nw *Network) IsPredefined() bool {
	return nw.Name == "bridge" || nw.Name == "null" || nw.Name == "host"
}
