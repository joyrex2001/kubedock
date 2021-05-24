package types

import (
	"time"
)

// Image describes the details of an image.
type Image struct {
	ID           string
	ShortID      string
	Name         string
	ExposedPorts map[string]interface{}
	Cmd          []string
	Created      time.Time
}
