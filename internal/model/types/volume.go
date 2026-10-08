package types

import (
	"regexp"
	"time"
)

// Volume describes the details of a named volume. Volumes are only
// administrative records; container mounts of type volume are not backed
// by storage.
type Volume struct {
	Name    string
	Labels  map[string]string
	Created time.Time
}

// Match will match given type with given key value pair.
func (vol *Volume) Match(typ string, key string, val string) (bool, error) {
	switch typ {
	case "name":
		return regexp.MatchString(key, vol.Name)
	case "label":
		v, ok := vol.Labels[key]
		if !ok {
			return false, nil
		}
		return val == "" || v == val, nil
	}
	return true, nil
}
