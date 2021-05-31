package filter

import (
	"encoding/json"
	"strings"
)

// FilterRequest is the json filter argument.
// Unfortunately, depending on which docker-compose or "docker compose" you
// run, the request may actually differ :-/ As "docker compose" doesn't
// really rely on the server-side filtering (it will filter at the client
// side as well), only the below type of request is supported for now...
type FilterRequest map[string][]string

// FilterMatcher is the interface for a Match method to test the filter.
type FilterMatcher interface {
	Match(string, string, string) bool
}

// instance is the instace of this filter object
type instance struct {
	filters map[string][]keyval
}

// keyval contains a key value pair for matching
type keyval struct {
	K string
	V string
}

// New will return a new filter instance
func New(f string) (*instance, error) {
	in := &instance{
		filters: map[string][]keyval{},
	}

	rq := FilterRequest{}
	if f != "" {
		if err := json.Unmarshal([]byte(f), &rq); err != nil {
			return in, err
		}
	}

	for typ, filtrs := range rq {
		if _, ok := in.filters[typ]; !ok {
			in.filters[typ] = []keyval{}
		}
		for _, f := range filtrs {
			flds := strings.Split(f, "=")
			if len(flds) != 2 {
				in.filters[typ] = append(in.filters[typ], keyval{flds[0], ""})
			} else {
				in.filters[typ] = append(in.filters[typ], keyval{flds[0], flds[1]})
			}
		}
	}
	return in, nil
}

// Match will call the matcher function and test if the object matches the
// given key values.
func (in *instance) Match(matcher FilterMatcher) bool {
	res := true
	for typ, filtrs := range in.filters {
		for _, f := range filtrs {
			if !matcher.Match(typ, f.K, f.V) {
				res = false
			}
		}
	}
	return res
}
