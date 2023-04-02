package types

import (
	"time"
)

// Network describes the details of a network.
type Network struct {
	ID      string
	ShortID string
	Name    string
	Labels  map[string]string
	Created time.Time
}

// IsPredefined will return if the network is a pre-defined system network.
func (nw *Network) IsPredefined() bool {
	return nw.Name == "bridge" || nw.Name == "null" || nw.Name == "host"
}

// Match will match given type with given key value pair.
func (nw *Network) Match(typ string, key string, val string) bool {
	if typ == "name" {
		return nw.Name == key
	}
	if typ != "label" {
		return true
	}
	v, ok := nw.Labels[key]
	if !ok {
		return false
	}
	return v == val
}
