package filter

import (
	"testing"
)

type matcher struct {
	res bool
}

func (m *matcher) Match(t string, k string, v string) (bool, error) {
	return m.res, nil
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
			suc:     true,
			match:   false,
		},
		{
			filter:  `{"label":{"com.docker.compose.project=timesheet":true}}`,
			matcher: &matcher{true},
			suc:     true,
			match:   true,
		},
		{
			filter:  `{"label":{"com.docker.compose.project=timesheet":true},"name":{"mycontainer":true}}`,
			matcher: &matcher{true},
			suc:     true,
			match:   true,
		},
		{
			filter:  `{"container":{"f577e780ec1756037235f0d5ba8081dfcdeb30327c75513f088953fa979b79b3":true},"type":{"container":true}}`,
			matcher: &matcher{true},
			suc:     true,
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
