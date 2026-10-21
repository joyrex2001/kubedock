package backend

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCommandWithEnv(t *testing.T) {
	tests := []struct {
		name    string
		command []string
		env     []string
		want    []string
	}{
		{
			name:    "no environment",
			command: []string{"pwd"},
			want:    []string{"pwd"},
		},
		{
			name:    "environment is passed as arguments",
			command: []string{"sh", "-c", "printf '%s' \"$DOCKER_HOST\""},
			env:     []string{"DOCKER_HOST=tcp://kubedock:2475", "SAMPLE=a b"},
			want:    []string{"env", "DOCKER_HOST=tcp://kubedock:2475", "SAMPLE=a b", "sh", "-c", "printf '%s' \"$DOCKER_HOST\""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := commandWithEnv(test.command, test.env); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("commandWithEnv() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestParseExecResponse(t *testing.T) {
	tests := []struct {
		in  error
		cod int
		suc bool
	}{
		{nil, 0, true},
		{fmt.Errorf("some generic error"), 0, false},
		{fmt.Errorf("command terminated with exit code 2"), 2, true},
	}

	for i, tst := range tests {
		kub := &instance{}
		cod, err := kub.parseExecResponse(tst.in)
		if cod != tst.cod {
			t.Errorf("failed test %d - expected %d, but got %d", i, tst.cod, cod)
		}
		if err != nil && tst.suc {
			t.Errorf("failed test %d - unexpected error: %s", i, err)
		}
		if err == nil && !tst.suc {
			t.Errorf("failed test %d - expected error, but succeeded instead", i)
		}
	}
}
