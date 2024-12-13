package backend

import (
	"testing"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestToKubernetesValue(t *testing.T) {
	tests := []struct {
		in    string
		key   string
		value string
		name  string
	}{
		{in: "__-abc", key: "abc", value: "abc", name: "abc"},
		{in: "/a/b/c", key: "a/b/c", value: "abc", name: "abc"},
		{
			in:    "StrategicMars",
			key:   "StrategicMars",
			value: "StrategicMars",
			name:  "StrategicMars",
		},
		{
			in:    "2107007e-b7c8-df23-18fb-6a6f79726578",
			key:   "2107007e-b7c8-df23-18fb-6a6f79726578",
			value: "2107007e-b7c8-df23-18fb-6a6f79726578",
			name:  "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:    "0123456789012345678901234567890123456789012345678901234567890123456789",
			key:   "012345678901234567890123456789012345678901234567890123456789012",
			value: "012345678901234567890123456789012345678901234567890123456789012",
			name:  "012345678901234567890123456789012345678901234567890123456789012",
		},
		{
			in:    "StrategicMars-",
			key:   "StrategicMars",
			value: "StrategicMars",
			name:  "StrategicMars",
		},
		{
			in:    "StrategicMars/-",
			key:   "StrategicMars",
			value: "StrategicMars",
			name:  "StrategicMars",
		},
		{
			in:    "2107007e-b7c8-df23-18fb-6a6f79726578",
			key:   "2107007e-b7c8-df23-18fb-6a6f79726578",
			value: "2107007e-b7c8-df23-18fb-6a6f79726578",
			name:  "2107007e-b7c8-df23-18fb-6a6f79726578",
		},
		{
			in:    "app.kubernetes.io/name",
			key:   "app.kubernetes.io/name",
			value: "app.kubernetes.ioname",
			name:  "appkubernetesioname",
		},
		{
			in:    "",
			key:   "",
			value: "",
			name:  "undef",
		},
	}

	for i, tst := range tests {
		kub := &instance{}
		key := kub.toKubernetesKey(tst.in)
		if key != tst.key {
			t.Errorf("failed test %d - expected key %s, but got %s", i, tst.key, key)
		}
		value := kub.toKubernetesValue(tst.in)
		if value != tst.value {
			t.Errorf("failed test %d - expected value %s, but got %s", i, tst.value, value)
		}
		name := kub.toKubernetesName(tst.in)
		if name != tst.name {
			t.Errorf("failed test %d - expected name %s, but got %s", i, tst.name, name)
		}
	}
}

func TestMapContainerTCPPorts(t *testing.T) {
	tests := []struct {
		in  *types.Container
		out map[int]int
	}{
		{
			in: &types.Container{ExposedPorts: map[string]interface{}{
				"303/tcp": 0,
				"909/tcp": 0,
			},
			},
		},
	}
	kub := &instance{}
	for j := 0; j < 100; j++ {
		for i, tst := range tests {
			err := kub.MapContainerTCPPorts(tst.in)
			if err != nil {
				t.Errorf("failed test %d/%d - unexpected error: %s", i, j, err)
			}
			m := map[int]int{}
			for p := range tst.in.MappedPorts {
				if p < 1024 {
					t.Errorf("failed test %d/%d - invalid random port %d", i, j, p)
					break
				}
				if _, ok := m[p]; ok {
					t.Errorf("failed test %d/%d - tandom port collision, port %d already provided", i, j, p)
					break
				}
			}
		}
	}
}

func TestMapContainerTCPPortsSkipBoundPorts(t *testing.T) {
	kub := &instance{}
	c := &types.Container{
		ExposedPorts: map[string]interface{}{
			"303/tcp": 0,
			"80/tcp":  0,
		},
		HostPorts: map[int]int{
			-303: 303,
			8080: 80,
		},
	}
	if err := kub.MapContainerTCPPorts(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := len(c.MappedPorts)
	if n != 1 {
		t.Errorf("expected 1 mapped port, but got %d", n)
	}
	for src, dst := range c.MappedPorts {
		if src == 0 {
			t.Errorf("expected non-zero source port, but got %d", src)
		}
		if dst != 303 {
			t.Errorf("expected destination port 303, but got %d", dst)
		}
	}
}
