package filter

import (
	"testing"
)

type matcher struct {
	res bool
}

func (m *matcher) Match(t, k, v string) bool {
	return m.res
}

func TestFilter(t *testing.T) {
	tests := []struct {
		filter  string
		matcher *matcher
		suc     bool
		match   bool
	}{
		{
			filter:  `{"label": ["com.docker.compose.project=timesheet", "com.docker.compose.oneoff=False"]}`,
			suc:     true,
			matcher: &matcher{false},
			match:   false,
		},
		{
			filter:  `{"label": ["com.docker.compose.project=timesheet", "com.docker.compose.oneoff=False"]}`,
			suc:     true,
			matcher: &matcher{true},
			match:   true,
		},
		{
			filter:  `{el": ["com.docker.compose.project=timesheet", "com.docker.compose.oneoff=False"]}`,
			matcher: &matcher{false},
			suc:     false,
			match:   true,
		},
		{
			filter:  `{"status": ["created", "exited"], "label": ["com.docker.compose.project=timesheet", "com.docker.compose.service=keycloak", "com.docker.compose.oneoff=False"]}`,
			matcher: &matcher{false},
			suc:     true,
			match:   false,
		},
		{
			filter:  `{"label":{"com.docker.compose.project=timesheet":true}}`,
			matcher: &matcher{false},
			suc:     false, // TODO: support this format (docker compose)
			match:   true,
		},
		{
			filter:  ``,
			matcher: &matcher{false},
			match:   true,
			suc:     true,
		},
	}

	for i, tst := range tests {
		filtr, err := New(tst.filter)
		if tst.suc && err != nil {
			t.Errorf("failed test %d - unexpected error %s", i, err)
		}
		if !tst.suc && err == nil {
			t.Errorf("failed test %d - expected error, but succeeded instead", i)
		}
		if filtr != nil {
			if filtr.Match(tst.matcher) != tst.match {
				t.Errorf("failed test %d - unexpected match", i)
			}
		}
	}
}