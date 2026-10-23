package backend

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCommandWithEnvAndWorkdir(t *testing.T) {
	tests := []struct {
		name       string
		command    []string
		env        []string
		workingDir string
		want       []string
	}{
		{
			name:    "no environment or working directory",
			command: []string{"pwd"},
			want:    []string{"pwd"},
		},
		{
			name:    "environment is passed as arguments",
			command: []string{"sh", "-c", "printf '%s' \"$DOCKER_HOST\""},
			env:     []string{"DOCKER_HOST=tcp://kubedock:2475", "SAMPLE=a b"},
			want:    []string{"env", "DOCKER_HOST=tcp://kubedock:2475", "SAMPLE=a b", "sh", "-c", "printf '%s' \"$DOCKER_HOST\""},
		},
		{
			name:       "working directory is passed as a shell argument",
			command:    []string{"pwd"},
			workingDir: "/workspace/repository with spaces",
			want:       []string{"sh", "-c", `cd "$1" && shift && exec "$@"`, "kubedock-exec", "/workspace/repository with spaces", "pwd"},
		},
		{
			name:       "environment and working directory are both preserved",
			command:    []string{"sh", "-c", "printf '%s' \"$DOCKER_HOST\""},
			env:        []string{"DOCKER_HOST=tcp://kubedock:2475"},
			workingDir: "/workspace/repository",
			want:       []string{"env", "DOCKER_HOST=tcp://kubedock:2475", "sh", "-c", `cd "$1" && shift && exec "$@"`, "kubedock-exec", "/workspace/repository", "sh", "-c", "printf '%s' \"$DOCKER_HOST\""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := commandWithEnvAndWorkdir(test.command, test.env, test.workingDir); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("commandWithEnvAndWorkdir() = %v, want %v", got, test.want)
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
