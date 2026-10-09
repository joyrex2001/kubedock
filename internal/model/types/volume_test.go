package types

import (
	"testing"
)

func TestVolumeMatch(t *testing.T) {
	vol := &Volume{Name: "proj_data", Labels: map[string]string{"com.docker.compose.project": "proj"}}
	tests := []struct {
		typ   string
		key   string
		val   string
		match bool
	}{
		{typ: "name", key: "proj_data", match: true},
		{typ: "name", key: "proj", match: true},
		{typ: "name", key: "other", match: false},
		{typ: "label", key: "com.docker.compose.project", val: "proj", match: true},
		{typ: "label", key: "com.docker.compose.project", val: "other", match: false},
		{typ: "label", key: "com.docker.compose.project", match: true},
		{typ: "label", key: "missing", match: false},
		{typ: "dangling", key: "true", match: true},
	}
	for i, tst := range tests {
		match, err := vol.Match(tst.typ, tst.key, tst.val)
		if err != nil {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if match != tst.match {
			t.Errorf("failed test %d - expected %t, but got %t", i, tst.match, match)
		}
	}
}
