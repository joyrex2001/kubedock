package filter

import (
	"encoding/json"
	"strings"
)

// Request is the json filter argument.
type Request map[string]map[string]bool

// Matcher is the interface for a Match method to test the filter.
type Matcher interface {
	Match(string, string, string) (bool, error)
}

// Filter is the instance of this filter object
type Filter struct {
	filters map[string][]keyval
}

// keyval contains a key value pair for matching
type keyval struct {
	K string
	V string
	P bool
}

// New will return a new filter instance
func New(f string) (*Filter, error) {
	in := &Filter{
		filters: map[string][]keyval{},
	}

	rq := Request{}
	if f != "" {
		if err := unmarshal(f, &rq); err != nil {
			return in, err
		}
	}

	for typ, filtrs := range rq {
		if _, ok := in.filters[typ]; !ok {
			in.filters[typ] = []keyval{}
		}
		for f, p := range filtrs {
			flds := strings.Split(f, "=")
			if len(flds) != 2 {
				in.filters[typ] = append(in.filters[typ], keyval{flds[0], "", p})
			} else {
				in.filters[typ] = append(in.filters[typ], keyval{flds[0], flds[1], p})
			}
		}
	}
	return in, nil
}

// unmarshal will unmarshal the given json to a Request type. Unfortunately,
// depending on which docker-compose or "docker compose" you run, the request
// may actually differ :-/ This method detects the format and marshalls either
// to the same Request format.
func unmarshal(dat string, rq *Request) error {
	if err := json.Unmarshal([]byte(dat), &rq); err == nil {
		return nil
	}

	// convert legacy format to new format...
	rql := map[string][]string{}
	if err := json.Unmarshal([]byte(dat), &rql); err != nil {
		return err
	}

	for typ, filtrs := range rql {
		(*rq)[typ] = map[string]bool{}
		for _, f := range filtrs {
			(*rq)[typ][f] = true
		}
	}

	return nil
}

// Match will call the matcher function and test if the object matches the
// given key values.
// To match moby we need to
// - OR the matching inside of each type filters
// - AND the result of the above for each type
// See https://github.com/moby/moby/blob/master/daemon/internal/filters/parse.go
func (in *Filter) Match(matcher Matcher) bool {
	for typ, filtrs := range in.filters {
		foundMatch := false
		for _, f := range filtrs {
			if isMatch, err := matcher.Match(typ, f.K, f.V); err != nil {
				continue // follows the moby pattern, ignore erroneous filters altogether
			} else if isMatch == f.P { // we found a match in this type filters
				foundMatch = true
				break // since we OR results of a type filters its safe to break here
			}
		}
		// if any type doesn't have a match return false immediately
		if !foundMatch {
			return false
		}
	}
	// all types had a match
	return true
}
